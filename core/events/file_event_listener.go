package events

import (
	"context"
	"raidline/ripple/core/graph"
	"raidline/ripple/pgk/logger"
	"strings"
)

type FileEventListener struct {
	graphQuerier graph.ProjectQuerier
}

func NewFileListener(pg graph.ProjectQuerier) (*FileEventListener, error) {
	return &FileEventListener{
		graphQuerier: pg,
	}, nil
}

// events will maybe be another type, because we need the file name to go to the graph
func (fl *FileEventListener) Listen(ctx context.Context, fileChanged <-chan string,
	outCh chan []string) {
	logger.Info("Ready and listening...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, cleaning up...")
			return
		case file, ok := <-fileChanged:
			if !ok {
				logger.Info("Consumer: Event channel closed, exiting...")
				return
			}

			logger.Info("Got event : [%s]...\n", file) // this is the filename (with extension)
			filename := extractFilename(file)

			logger.Info("File : [%s] impacts: \n", filename)

			if fl.graphQuerier.Exists(filename) {
				impacts := fl.graphQuerier.FindAllWithEdge(filename)
				outCh <- impacts
			} else {
				logger.Info("The file we received update from is not in the graph. \n")
			}
		}
	}
}

func extractFilename(file string) string {
	splitted := strings.Split(file, "/")                    // /path/to/file/resources/test-files/InspectionZoneService.java
	return strings.Split(splitted[len(splitted)-1], ".")[0] // file.java, file.go, file.model.ts -> just "file"
}
