package chat

import (
	"raidline/ripple/core/graph"
	"time"
)

//todo: not sure what to do with this class for now

type GraphQuerier struct {
	pg graph.ProjectQuerier
}

func (g *GraphQuerier) Execute(query string) (string, error) { //not sure what to do with the error here
	time.Sleep(1500 * time.Millisecond)
	return "New message....", nil
}

func NewQuerier(pg graph.ProjectQuerier) *GraphQuerier {
	return &GraphQuerier{
		pg: pg,
	}
}
