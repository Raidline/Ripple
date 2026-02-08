package graph

import (
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
)

var packageBreakerSimbols = map[languages.Language]string{
	languages.JAVA: ".",
	languages.TS:   "/",
}

type ProjectGraphAggregator struct {
	Graph *model.ProjectGraph
}

func Create() *ProjectGraphAggregator {
	return &ProjectGraphAggregator{
		Graph: &model.ProjectGraph{
			Vertices: make(map[string]*model.GraphVertice, 0),
		},
	}
}

func (agg *ProjectGraphAggregator) Aggregate(files []*pgk.FileScan, wantedLang languages.Language) error {

	if files == nil {
		return errors.NewEmptySequenceError("files sequence")
	}

	fileAnalyser, e := languages.GetAnalyser(wantedLang)

	if e != nil {
		return e
	}

	fileNameToFileScan := make(map[string]*pgk.FileScan, 0)

	for _, file := range files {
		fileNameToFileScan[file.Name] = file
	}

	agg.createProjectGraph(fileNameToFileScan, fileAnalyser, wantedLang)

	return nil
}

func (agg *ProjectGraphAggregator) createProjectGraph(
	fileNameToFileScan map[string]*pgk.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) error {

	seen := make(map[string]bool, 0)

	for _, file := range fileNameToFileScan {
		_, e := agg.createGraphForFile(file, seen, fileNameToFileScan, fileAnalyser, wantedLang)

		if e != nil {
			return e
		}
	}

	return nil
}

func (agg *ProjectGraphAggregator) createGraphForFile(fileScan *pgk.FileScan,
	seen map[string]bool,
	fileNameToFileScan map[string]*pgk.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) (*model.GraphVertice, error) {

	if seen[fileScan.Name] {
		return nil, nil
	}

	if fileScan == nil {
		panic("Scan of the file cannot be empty")
	}

	fileGraph, fileGErr := languages.BuildFileGraph(fileScan, fileAnalyser)

	if fileGErr != nil {
		return nil, fileGErr
	}

	if vertice, ok := agg.Graph.Vertices[fileGraph.ClassName]; ok {
		// this might already exist but as a from dependency, now we need to add the to's
		agg.connectEdgesToVertice(vertice, fileNameToFileScan, fileGraph, func(fileScan *pgk.FileScan) (*model.GraphVertice, error) {
			return agg.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return vertice, nil
	} else {
		v := &model.GraphVertice{
			Node:  fileGraph,
			Edges: make([]model.GraphEdge, 0),
		}
		// todo(the fields and method info to get the weight of each import)
		agg.Graph.Vertices[fileGraph.ClassName] = v

		//Edges will be added by memory reference (i hope?)
		agg.connectEdgesToVertice(v, fileNameToFileScan, fileGraph, func(fileScan *pgk.FileScan) (*model.GraphVertice, error) {
			return agg.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return v, nil
	}
}

func (agg *ProjectGraphAggregator) connectEdgesToVertice(vert *model.GraphVertice,
	fileNameToFileScan map[string]*pgk.FileScan,
	fileGraph *model.ClassGraph, cb func(fileScan *pgk.FileScan) (*model.GraphVertice, error)) error {

	fieldToDependency := make(map[string]*pgk.FileScan, 0)

	for _, f := range fileGraph.Fields {
		if v, ok := fileNameToFileScan[f.Type]; ok {
			fieldToDependency[f.Type] = v
		}
	}

	for name, scan := range fieldToDependency {

		if vertice, ok := agg.Graph.Vertices[name]; ok {
			vert.Edges = append(vert.Edges, model.GraphEdge{
				To:     vertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			})
		} else {
			//build the file
			newVertice, e := cb(scan)

			if e != nil {
				return e
			}

			vert.Edges = append(vert.Edges, model.GraphEdge{
				To:     newVertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			})
		}
	}

	return nil
}
