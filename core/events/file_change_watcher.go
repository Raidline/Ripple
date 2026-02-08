package events

import (
	"context"
	"log"
	"os"
	"raidline/ripple/core/graph"
	"raidline/ripple/errors"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	fsWatcher *fsnotify.Watcher
}

func NewWatcher(pg *graph.ProjectGraphAggregator) (*FileWatcher, error) {
	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		fsWatcher: fswatcher,
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
func (f *FileWatcher) Watch(ctx context.Context, dirs []string) (<-chan string, error) {

	if len(dirs) == 0 {
		return nil, errors.NewEmptySequenceError("directory paths")
	}

	for _, dir := range dirs {
		if err := f.fsWatcher.Add(dir); err != nil {
			return nil, err
		}
	}

	outCh := make(chan string, 10)

	go func() {
		defer close(outCh)        // Close channel when loop exits
		defer f.fsWatcher.Close() // Clean up fsnotify

		for {
			select {
			case <-ctx.Done():
				return // Exit on context cancellation

			case event, ok := <-f.fsWatcher.Events:
				if !ok {
					return
				}

				if strings.Contains(event.Name, ".git") || strings.Contains(event.Name, ".node") || strings.HasSuffix(event.Name, "~") {
					continue
				}

				// Handle New Directories (Recursive Watch)
				if event.Op&fsnotify.Create == fsnotify.Create {
					info, err := os.Stat(event.Name)
					if err == nil && info.IsDir() {
						f.fsWatcher.Add(event.Name)
					}
				}

				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					continue
				}

				if hasFileNotChanged(event.Name) {
					continue
				}

				select {
				case outCh <- event.Name:
				case <-ctx.Done():
					return
				}

			case err, ok := <-f.fsWatcher.Errors:
				if !ok {
					return
				}
				log.Println("Watcher error:", err)
			}
		}
	}()

	// 4. Return immediately
	return outCh, nil
}

func hasFileNotChanged(file string) bool {
	// Run git diff-index --quiet HEAD file.extensions to make sure the change wasn't just a "touch" (where timestamp changed but content didn't).
	return true
}
