package core

import (
	"raidline/ripple/core/languages"
	"raidline/ripple/core/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
)

type ProjectGraphAggregator struct {
	Graph *model.ProjectGraph
}

func Create() *ProjectGraphAggregator {
	return &ProjectGraphAggregator{
		Graph: &model.ProjectGraph{
			Vertices: make(map[string]model.GraphVertice, 0),
		},
	}
}

func (agg *ProjectGraphAggregator) aggregate(rootDir string, lang string) error {

	var languageErr error
	var wantedLang languages.Language

	if lang == string(languages.JAVA) {
		wantedLang = languages.JAVA
	} else if lang == string(languages.TS) {
		wantedLang = languages.TS
	} else {
		languageErr = errors.NewLanguageNotSupportedError(lang)
	}

	if languageErr != nil {
		return languageErr
	}

	files, err := pgk.CreepDir(rootDir)

	if err != nil {
		return err
	}

	for file := range files {
		fileAnalyser, e := languages.GetAnalyser(wantedLang)

		if e != nil {
			return e
		}

		fileGraph, fileGErr := languages.BuildFileGraph(file, fileAnalyser)

		if fileGErr != nil {
			return fileGErr
		}

		agg.appendToCurrentGraph(fileGraph)
	}

	return nil
}

func (agg *ProjectGraphAggregator) appendToCurrentGraph(fileGraph *model.ClassGraph) {
	if vertice, ok := agg.Graph.Vertices[fileGraph.ClassName]; ok {
		connectEdges(vertice, fileGraph)
	} else {
		v := model.GraphVertice{}
		v.Node = fileGraph
		connectEdges(v, fileGraph)
		// todo(the fields and method info to get the weight of each import)
		agg.Graph.Vertices[fileGraph.ClassName] = v
	}
}

func connectEdges(v1 model.GraphVertice, fileGraph *model.ClassGraph) {

}
