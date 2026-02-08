package events

import (
	"context"
	"fmt"
	"raidline/ripple/core/graph"
	"time"
)

type FileEventListener struct {
	graph *graph.ProjectGraphAggregator
}

func NewFileListener(pg *graph.ProjectGraphAggregator) (*FileEventListener, error) {
	return &FileEventListener{
		graph: pg,
	}, nil
}

//events will maybe be another type, because we need the file name to go to the graph

func (fl *FileEventListener) Listen(ctx context.Context, events <-chan string) {
	fmt.Println("Consumer: Ready and listening...")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Consumer: Context cancelled, cleaning up...")
			return
		case event, ok := <-events:
			if !ok {
				fmt.Println("Consumer: Event channel closed, exiting...")
				return
			}

			fmt.Printf("Consumer: Analyzing [%s]...\n", event)
			time.Sleep(500 * time.Millisecond) // Simulating work
		}
	}
}
