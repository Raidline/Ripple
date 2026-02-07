package graph

import (
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/pgk"
	"testing"
)

func TestAggregate(t *testing.T) {

	agg := Create()

	creepResult, _ := pgk.CreepDir("../../resources/test-files")

	e := agg.Aggregate(creepResult.Files, languages.JAVA)

	if e != nil {
		t.Errorf("Should be able to construct Graph : %s", e.Error())
		t.FailNow()
	}

	if len(agg.Graph.Vertices) != 3 {
		t.Errorf("Should be able to have 3 Vertices in the graph : %d", len(agg.Graph.Vertices))
		t.FailNow()
	}

	if v, ok := agg.Graph.Vertices["InspectionZoneService"]; ok {
		if len(v.Edges) != 2 {
			t.Errorf("Should be able to have 2 Edges in the graph for InspectionZoneService: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to InspectionZoneService as a vertice")
		t.FailNow()
	}

	if v, ok := agg.Graph.Vertices["SecondService"]; ok {
		if len(v.Edges) != 2 {
			t.Errorf("Should be able to have 2 Edge in the graph for SecondService: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to SecondService as a vertice")
		t.FailNow()
	}

	if v, ok := agg.Graph.Vertices["InspectionZoneRepository"]; ok {
		if len(v.Edges) != 0 {
			t.Errorf("Should be able to have 0 Edge in the graph for InspectionZoneRepository: %d", len(v.Edges))
			t.FailNow()
		}
	} else {
		t.Errorf("Should be able to InspectionZoneRepository as a vertice")
		t.FailNow()
	}

}
