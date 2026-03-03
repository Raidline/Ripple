package services

import (
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
)

type projectQuerier struct {
	state *domain.StateCoordinator
}

func NewQuerier(state *domain.StateCoordinator) ports.ProjectGraphQuerier {
	return &projectQuerier{
		state: state,
	}
}

func (pg *projectQuerier) FindAllWithEdge(filename string) []string {
	if len(pg.state.Graph.Vertices) == 0 {
		return make([]string, 0)
	}
	target, exists := pg.state.Graph.Vertices[filename]
	if !exists {
		return nil
	}

	var dependents []string
	for _, edge := range target.InboundEdges {
		dependents = append(dependents, edge.From.Node.ClassName)
	}
	return dependents
}

func (pg *projectQuerier) Exists(filename string) bool {
	_, ok := pg.state.Graph.Vertices[filename]

	return ok
}
