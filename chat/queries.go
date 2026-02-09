package chat

import "raidline/ripple/core/graph"

type GraphQuerier struct {
	pg *graph.ProjectGraphAggregator
}

func (g *GraphQuerier) Execute(query string) (string, error) { //not sure what to do with the error here
	return "New message....", nil
}

func NewQuerier(pg *graph.ProjectGraphAggregator) *GraphQuerier {
	return &GraphQuerier{
		pg: pg,
	}
}
