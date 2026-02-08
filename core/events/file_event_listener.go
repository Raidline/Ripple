package events

import (
	"context"
	"fmt"
	"raidline/ripple/core/graph"
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
	fmt.Println("Ready and listening...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context cancelled, cleaning up...")
			return
		case file, ok := <-fileChanged:
			if !ok {
				fmt.Println("Consumer: Event channel closed, exiting...")
				return
			}

			fmt.Printf("Got event : [%s]...\n", file) // this is the filename (with extension)
			filename := extractFilename(file)

			if fileVertice, ok := fl.projectGraph.Graph.Vertices[filename]; ok {
				fmt.Printf("File : [%s] impacts: \n", filename)
				// verify if here we want to do a BFS or DFS to know the real impact of the file.
				// This is just the direct impacts, which is fine, but we can (if requested) make a more deep find
				for _, edge := range fileVertice.Edges {
					fmt.Printf("[%s] \n", edge.To.Node.ClassName)
				}
			} else {
				// this can be a case where the file was added. In this case we just update the projectGraph
				// todo: send update to the graph here, for now panic
				panic(fmt.Errorf("The file we received update from is not in the graph"))
			}
		}
	}
}

func extractFilename(file string) string {
	return strings.Split(file, ".")[0] // file.java, file.go, file.model.ts -> just "file"
}
