package watcher

import (
	"raidline/ripple/core/graph"
	"raidline/ripple/errors"
)

// (kqueue on macOS, inotify on Linux, ReadDirectoryChangesW on Windows).
// Use a library like fsnotify.
// fsnotify does not support recursive, need to implement mannualy.
// It seems we can add the path we are going through to the main watcher. This creates a circular dependency between the project builder and the FileWatcher as things are
// DirCrepeer will be present a level above and project.Aggregate will receive the sequence of files already, instead of building them there
// CreepDir will need to return a more well-defined struct with the *FileScan and all the dirs as a []string to be added to the watchers
// !Should not receive the watcher because we need to defer the close of watcher after creation, and this file should be the one responsible for it!
// This struct will then receive the path[] to create the watchers and the graph
//
// Code example for the watcher:
//watcher, err := fsnotify.NewWatcher()
// if err != nil {
// 	log.Fatal(err)
// }
// defer watcher.Close()

// // The root directory you want to watch
// rootPath := "./"

// // 1. Initial Walk: Add all existing directories
// err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
// 	if err != nil {
// 		return err
// 	}
// 	// Don't watch the .git folder! It changes constantly and will melt your CPU.
// 	if info.IsDir() && !strings.Contains(path, ".git") {
// 		return watcher.Add(path)
// 	}
// 	return nil
// })
// if err != nil {
// 	log.Fatal(err)
// }

// done := make(chan bool)

// go func() {
// 	for {
// 		select {
// 		case event, ok := <-watcher.Events:
// 			if !ok {
// 				return
// 			}

// 			// IGNORE git operations and temporary files
// 			if strings.Contains(event.Name, ".git") || strings.HasSuffix(event.Name, "~") {
// 				continue
// 			}

// 			log.Println("event:", event)

// 			// 2. Dynamic Watch: Did we just create a new directory?
// 			if event.Op&fsnotify.Create == fsnotify.Create {
// 				info, err := os.Stat(event.Name)
// 				if err == nil && info.IsDir() {
// 					log.Printf("New directory detected: %s. Adding watcher...", event.Name)
// 					watcher.Add(event.Name)
// 				}
// 			}

// 			// YOUR LOGIC HERE:
// 			// if strings.HasSuffix(event.Name, ".go") {
// 			//     RunMyTool()
// 			// }

// 		case err, ok := <-watcher.Errors:
// 			if !ok {
// 				return
// 			}
// 			log.Println("error:", err)
// 		}
// 	}
// }()

// <-done
//

type FileWatcher struct {
	projectGraph *graph.ProjectGraphAggregator
}

func NewWatcher(pg *graph.ProjectGraphAggregator) (*FileWatcher, error) {
	return &FileWatcher{
		projectGraph: pg,
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
func (f *FileWatcher) Watch(dirs []string) error {

	if len(dirs) == 0 {
		return errors.NewEmptySequenceError("directory paths")
	}

	return nil
}
