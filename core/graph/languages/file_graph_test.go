package languages

import (
	"raidline/ripple/pgk"
	"testing"
)

func TestBuildFileGraph(t *testing.T) {

	creepRes, err := pgk.CreepDir("../../../resources/test-files")

	if err != nil {
		t.Errorf("Should be able to read the test resources, %s", err.Error())
		t.FailNow()
	}

	var foundFile *pgk.FileScan

	for _, f := range creepRes.Files {
		if f.Name == "InspectionZoneService" {
			foundFile = f
		}
	}

	analyser, e := GetAnalyser(JAVA)

	if e != nil {
		t.Errorf("Should be able to create analyser, %s", e.Error())
		t.FailNow()
	}

	graph, ge := BuildFileGraph(foundFile, analyser)

	if ge != nil {
		t.Errorf("Should be able to construct a graph for the file : %s -> [%s]", foundFile.Name, ge.Error())
		t.FailNow()
	}

	if graph == nil {
		t.Errorf("Graph cannot be nil for the file : %s", foundFile.Name)
		t.FailNow()
	}

	if graph.ClassName != "InspectionZoneService" {
		t.Errorf("File should have name InspectionZoneService, not : [%s]", graph.ClassName)
		t.FailNow()
	}

	if len(graph.Imports) != 4 {
		t.Errorf("File should have 4 non std imports, not : [%d]", len(graph.Imports))
		t.FailNow()
	}

	if len(graph.Fields) != 2 {
		t.Errorf("Fields should have 2 fields, not : [%d]", len(graph.Fields))
		t.FailNow()
	}

	repoField := graph.Fields[0]

	if repoField.Name != "repository" {
		t.Errorf("First field should be the repository field, not : [%s]", repoField.Name)
		t.FailNow()
	}

	if repoField.Type != "InspectionZoneRepository" {
		t.Errorf("First field should be the InspectionZoneRepository type, not : [%s]", repoField.Type)
		t.FailNow()
	}

	serviceField := graph.Fields[1]

	if serviceField.Name != "s" {
		t.Errorf("First field should be the s field, not : [%s]", serviceField.Name)
		t.FailNow()
	}

	if serviceField.Type != "SecondService" {
		t.Errorf("First field should be the SecondService type, not : [%s]", serviceField.Type)
		t.FailNow()
	}

	if len(graph.Methods) != 3 {
		t.Errorf("Should have 3 files, not [%d]", len(graph.Methods))
		t.FailNow()
	}

	//updateZoneReferences assertion
	m0 := graph.Methods[0]
	if m0.Name != "updateZoneReferences" {
		t.Errorf("Method 0 name mismatch: expected [updateZoneReferences], got [%s]", m0.Name)
	}
	if m0.ReturnType != "void" {
		t.Errorf("Method 0 return type mismatch: expected [void], got [%s]", m0.ReturnType)
	}
	if len(m0.Params) != 1 {
		t.Errorf("Method 0 should have 1 param, got [%d]", len(m0.Params))
	} else {
		if m0.Params[0].Type != "List<TechnologyStructureZone>" {
			t.Errorf("Param type mismatch: expected [List<TechnologyStructureZone>], got [%s]", m0.Params[0].Type)
		}
	}
	// Checking the call to the repository
	foundPersist := false
	for _, call := range m0.Calls {
		if call.Target == "this.repository" && call.Method == "persistOrUpdate" {
			foundPersist = true
			break
		}
	}
	if !foundPersist {
		t.Errorf("Method 0 missing call to this.repository.persistOrUpdate")
	}

	// findOrCreate assertion
	m1 := graph.Methods[1]
	if m1.Name != "findOrCreate" {
		t.Errorf("Method 1 name mismatch: expected [findOrCreate], got [%s]", m1.Name)
	}
	if m1.ReturnType != "InspectionZone" {
		t.Errorf("Method 1 return type mismatch: expected [InspectionZone], got [%s]", m1.ReturnType)
	}

	// Params check
	if len(m1.Params) != 2 {
		t.Errorf("Method 1 should have 2 params, got [%d]", len(m1.Params))
	} else {
		if m1.Params[1].Type != "Map<String, InspectionZone>" {
			t.Errorf("Param 1 type mismatch: expected [Map<String, InspectionZone>], got [%s]", m1.Params[1].Type)
		}
	}

	// Local variable calls (the 'zone' object)
	expectedCalls := []string{"containsKey", "get", "setId", "setActive", "setInspectionArea"}
	for _, expected := range expectedCalls {
		found := false
		for _, call := range m1.Calls {
			if call.Method == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Method 1 missing expected call to [%s]", expected)
		}
	}

	// getZonesById assertion
	m2 := graph.Methods[2]
	if m2.Name != "getZonesById" {
		t.Errorf("Method 2 name mismatch: expected [getZonesById], got [%s]", m2.Name)
	}
	if m2.ReturnType != "Map<String, InspectionZone>" {
		t.Errorf("Method 2 return type mismatch: expected [Map<String, InspectionZone>], got [%s]", m2.ReturnType)
	}

	streamCalls := map[string]bool{
		"findAllByIds": false,
		"stream":       false,
		"map":          false,
		"toList":       false,
		"collect":      false,
	}

	for _, call := range m2.Calls {
		if _, ok := streamCalls[call.Method]; ok {
			streamCalls[call.Method] = true
		}
	}

	for method, found := range streamCalls {
		if !found {
			t.Errorf("Method 2 missing call in stream chain: [%s]", method)
		}
	}
}
