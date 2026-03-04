package services

import (
	"context"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/domain"
	"sort"
	"testing"
)

func TestFindAllWithEdge(t *testing.T) {
	vAuth := &model.GraphVertice{Edges: []model.GraphEdge{}}
	vUser := &model.GraphVertice{Edges: []model.GraphEdge{}}
	vUtil := &model.GraphVertice{Edges: []model.GraphEdge{}}
	vMain := &model.GraphVertice{Edges: []model.GraphEdge{}}

	vAuth.Edges = append(vAuth.Edges, model.GraphEdge{From: vAuth, To: vUtil})

	vUser.Edges = append(vUser.Edges, model.GraphEdge{From: vUser, To: vUtil})

	vMain.Edges = append(vMain.Edges, model.GraphEdge{From: vMain, To: vAuth})
	vMain.Edges = append(vMain.Edges, model.GraphEdge{From: vMain, To: vUser})

	state := domain.NewStateCoordinator(context.Background())

	state.Graph.Vertices = map[string]*model.GraphVertice{
		"AuthService": vAuth,
		"UserService": vUser,
		"Utils":       vUtil,
		"Main":        vMain,
	}

	pg := NewQuerier(state)

	tests := []struct {
		name     string
		filename string   // The file we are investigating
		want     []string // Who depends on this file?
	}{
		{
			name:     "Find who depends on Utils (Multiple Dependents)",
			filename: "Utils",
			want:     []string{"AuthService", "UserService"},
		},
		{
			name:     "Find who depends on AuthService (Single Dependent)",
			filename: "AuthService",
			want:     []string{"Main"},
		},
		{
			name:     "Find who depends on Main (No Dependents)",
			filename: "Main",
			want:     []string{}, // Nobody imports Main
		},
		{
			name:     "Non-existent file",
			filename: "GhostFile",
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pg.FindAllWithEdge(tt.filename)

			// Sort both slices because map iteration order is random in Go.
			// Without sorting, ["A", "B"] might fail against ["B", "A"] even if correct.
			sort.Strings(got)
			sort.Strings(tt.want)

			if !slicesEqual(got, tt.want) {
				t.Errorf("FindAllWithEdge(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// Helper to compare slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
