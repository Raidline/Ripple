package core

import (
	"raidline/ripple/core/languages"
	"raidline/ripple/core/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
	"strings"
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

	fileAnalyser, e := languages.GetAnalyser(wantedLang)

	if e != nil {
		return e
	}

	for file := range files {

		fileGraph, fileGErr := languages.BuildFileGraph(file, fileAnalyser)

		if fileGErr != nil {
			return fileGErr
		}

		agg.appendToCurrentGraph(wantedLang, fileGraph)
	}

	return nil
}

func (agg *ProjectGraphAggregator) appendToCurrentGraph(lang languages.Language, fileGraph *model.ClassGraph) {
	if vertice, ok := agg.Graph.Vertices[fileGraph.ClassName]; !ok {
		agg.connectEdgesToVertice(lang, vertice, fileGraph) // this might already exist but as a from dependency, now we need to add the to's
	} else {
		v := model.GraphVertice{}
		v.Node = fileGraph
		agg.connectEdgesToVertice(lang, v, fileGraph)
		// todo(the fields and method info to get the weight of each import)
		agg.Graph.Vertices[fileGraph.ClassName] = v
	}
}

func (agg *ProjectGraphAggregator) connectEdgesToVertice(lang languages.Language,
	v1 model.GraphVertice, fileGraph *model.ClassGraph) error {

	projectImports := make([]string, len(fileGraph.Imports)) // all the project imports ready to go
	if breakSimbol, ok := packageBreakerSimbols[lang]; ok {
		for _, v := range fileGraph.Imports {
			splitted := strings.Split(v, breakSimbol)

			impName := splitted[len(splitted)-1] //here we trust that this is a project dependency
			projectImports = append(projectImports, impName)
		}
	} else {
		return errors.NewLanguageNotSupportedError(string(lang))
	}

	for _, v := range projectImports {
		if vertice, ok := agg.Graph.Vertices[v]; ok {
			v1.Edges = append(v1.Edges, model.GraphEdge{
				To:     vertice.Node,
				From:   fileGraph,
				Weight: 0, // should be the count of times this import is used
			})
		} else {
			//todo: build the file, we need to call this in a recursive manner, refactor the method after. break it into smaller pieces
		}
	}

	return nil
}
