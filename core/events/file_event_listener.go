package events

import (
	"context"
	"raidline/ripple/core/graph"
	"raidline/ripple/pgk/logger"
	"strings"
)

type FileEventListener struct {
	projectGraph *graph.ProjectGraphAggregator
}

func NewFileListener(pg *graph.ProjectGraphAggregator) (*FileEventListener, error) {
	return &FileEventListener{
		projectGraph: pg,
	}, nil
}

//events will maybe be another type, because we need the file name to go to the graph

func (fl *FileEventListener) Listen(ctx context.Context, fileChanged <-chan string) { // for now we assume this is the file name
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

			if fileVertice, ok := fl.projectGraph.Graph.Vertices[filename]; ok {
				logger.Info("File : [%s] impacts: \n", filename)
				// verify if here we want to do a BFS or DFS to know the real impact of the file.
				// This is just the direct impacts, which is fine, but we can (if requested) make a more deep find
				for _, edge := range fileVertice.Edges {
					logger.Info("[%s] \n", edge.To.Node.ClassName)
				}
			} else {
				// this can be a case where the file was added. In this case we just update the projectGraph
				// todo: send update to the graph here
				// When the Watcher updates a file, it should create a totally new GraphVertice object and replace the old one in the map.
				//The TUI will keep holding the "old" version it was reading (which is safe, because that old slice isn't being modified anymore).
				//The Map will now point to the "new" version.
				logger.Info("The file we received update from is not in the graph. \n")
			}
		}
	}
}

func extractFilename(file string) string {
	return strings.Split(file, ".")[0] // file.java, file.go, file.model.ts -> just "file"
}
