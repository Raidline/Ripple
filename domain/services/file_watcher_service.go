package services

import (
	"math"
	"os"
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
	"raidline/ripple/errors"
	"raidline/ripple/pgk/logger"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type fileWatcher struct {
	fsWatcher  *fsnotify.Watcher
	stateCoord *domain.StateCoordinator
}

func NewWatcher(state *domain.StateCoordinator) (ports.FileWatcherPort, error) {
	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fileWatcher{
		fsWatcher:  fswatcher,
		stateCoord: state,
	}, nil
}

// Use the operating system's native file system events
//
//	Watch: The OS tells your tool "Event: Write on file.extension".
//	React: Ripple receives the event instantly.
//	Verify: Run git diff-index --quiet HEAD file.extensions to make sure the change wasn't just a "touch" (where timestamp changed but content didn't).
//
// Creates the watcher based on the paths found of the project
// Listens to file changes and and calls the project graph to get the ripple effects
//
// Returns a channel that is triggered when a change happens to a certain file, or error
func (f *fileWatcher) Watch(dirs []string) (<-chan string, error) {
	outCh := make(chan string, 10)

	var (
		// Wait 100ms for new events; each new event resets the timer.
		waitFor = 100 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)
	)

	if len(dirs) == 0 {
		return nil, errors.NewEmptySequenceError("directory paths")
	}

	for _, dir := range dirs {
		if err := f.fsWatcher.Add(dir); err != nil {
			return nil, err
		}
	}

	go func() {
		defer close(outCh)        // Close channel when loop exits
		defer f.fsWatcher.Close() // Clean up fsnotify

		for {
			select {
			case <-f.stateCoord.Ctx.Done():
				return

			case event, ok := <-f.fsWatcher.Events:
				if !ok {
					return
				}

				if strings.Contains(event.Name, ".git") || strings.Contains(event.Name, ".node") || strings.HasSuffix(event.Name, "~") {
					continue
				}

				if event.Op.Has(fsnotify.Rename) {
					logger.Debug("file [%s] was renamed, we received this old path, Created event should be received next", event.Name)
					// in here we should store this rename and wait for the respective create event
					// in the create event we should take all the references of the old path and update them to the new path
					continue
				}

				if event.Op.Has(fsnotify.Create) {
					// this can be a case where the file was added. In this case we just update the projectGraph
					// todo: send update to the graph here
					// When the Watcher updates a file, it should create a totally new GraphVertice object and replace the old one in the map.
					//The TUI will keep holding the "old" version it was reading (which is safe, because that old slice isn't being modified anymore).
					//The Map will now point to the "new" version.
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						f.fsWatcher.Add(event.Name)
					}
				}

				if !event.Op.Has(fsnotify.Write) {
					continue
				}

				// The pathname was written to; this does *not* mean the write has finished,
				// and a write can be followed by more writes.
				//
				// we need to wait for the last write on this file.
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					continue
				}

				// Get timer.
				mu.Lock()
				t, ok := timers[event.Name]
				mu.Unlock()

				if !ok {
					t = time.AfterFunc(math.MaxInt64, func() { f.writeEvent(&mu, timers, event, outCh) })
					t.Stop()

					mu.Lock()
					timers[event.Name] = t
					mu.Unlock()
				}

				t.Reset(waitFor)

			case err, ok := <-f.fsWatcher.Errors:
				if !ok {
					return
				}
				logger.Error("Watcher error: [%s]", err.Error())
				return
			}
		}
	}()

	return outCh, nil
}

func (f *fileWatcher) writeEvent(mu *sync.Mutex, timers map[string]*time.Timer, e fsnotify.Event, outCh chan string) {

	var (
		extractFilename = func(file string) string {
			splitted := strings.Split(file, "/")                    // /path/to/file/resources/test-files/InspectionZoneService.java
			return strings.Split(splitted[len(splitted)-1], ".")[0] // file.java, file.go, file.model.ts -> just "file"
		}
	)

	mu.Lock()
	delete(timers, e.Name)
	mu.Unlock()

	filename := extractFilename(e.Name)

	outCh <- filename
}
