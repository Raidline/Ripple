package graph

import (
	"raidline/ripple/core/graph/creeper"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
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

	pg := &ProjectGraph{
		graph: &model.ProjectGraph{
			Vertices: map[string]*model.GraphVertice{
				"AuthService": vAuth,
				"UserService": vUser,
				"Utils":       vUtil,
				"Main":        vMain,
			},
		},
	}

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

func TestAggregate(t *testing.T) {

	agg := CreateProjectGraph()

	creepResult, _ := creeper.CreepDir("../../resources/test-files")

	e := agg.Aggregate(creepResult.Files, languages.JAVA)

	if e != nil {
		t.Errorf("Should be able to construct Graph : %s", e.Error())
		t.FailNow()
	}

	if len(agg.graph.Vertices) != 3 {
		t.Errorf("Should be able to have 3 Vertices in the graph : %d", len(agg.graph.Vertices))
		t.FailNow()
	}

	if v, ok := agg.graph.Vertices["InspectionZoneService"]; ok {
		if len(v.Edges) != 2 {
			t.Errorf("Should be able to have 2 Edges in the graph for InspectionZoneService: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to InspectionZoneService as a vertice")
		t.FailNow()
	}

	if v, ok := agg.graph.Vertices["SecondService"]; ok {
		if len(v.Edges) != 2 {
			t.Errorf("Should be able to have 2 Edge in the graph for SecondService: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to SecondService as a vertice")
		t.FailNow()
	}

	if v, ok := agg.graph.Vertices["InspectionZoneRepository"]; ok {
		if len(v.Edges) != 0 {
			t.Errorf("Should be able to have 0 Edge in the graph for InspectionZoneRepository: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to InspectionZoneRepository as a vertice")
		t.FailNow()
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
