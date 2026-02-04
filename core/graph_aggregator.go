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

func (agg *ProjectGraphAggregator) Aggregate(rootDir string, lang string) error {

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

	fileNameToFileScan := make(map[string]*pgk.FileScan, 0)

	for file := range files {
		fileNameToFileScan[file.Name] = file
	}

	agg.createProjectGraph(fileNameToFileScan, fileAnalyser, wantedLang)

	return nil
}

func (agg *ProjectGraphAggregator) createProjectGraph(
	fileToScan map[string]*pgk.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) error {

	for _, file := range fileToScan {
		agg.createGraphForFile(file, fileToScan, fileAnalyser, wantedLang)
	}

	return nil
}

func (agg *ProjectGraphAggregator) createGraphForFile(fileName *pgk.FileScan, fileToScan map[string]*pgk.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) (*model.ClassGraph, error) {

	fileGraph, fileGErr := languages.BuildFileGraph(fileName, fileAnalyser)

	if fileGErr != nil {
		return nil, fileGErr
	}

	if vertice, ok := agg.Graph.Vertices[fileGraph.ClassName]; !ok {
		// this might already exist but as a from dependency, now we need to add the to's
		agg.connectEdgesToVertice(wantedLang, vertice, fileGraph, func(depName string) (*model.ClassGraph, error) {
			return agg.createGraphForFile(fileToScan[depName], fileToScan, fileAnalyser, wantedLang)
		})
	} else {
		v := model.GraphVertice{}
		v.Node = fileGraph
		agg.connectEdgesToVertice(wantedLang, v, fileGraph, func(depName string) (*model.ClassGraph, error) {
			return agg.createGraphForFile(fileToScan[depName], fileToScan, fileAnalyser, wantedLang)
		})
		// todo(the fields and method info to get the weight of each import)
		agg.Graph.Vertices[fileGraph.ClassName] = v
	}

	return fileGraph, nil
}

func (agg *ProjectGraphAggregator) connectEdgesToVertice(lang languages.Language,
	v1 model.GraphVertice, fileGraph *model.ClassGraph, cb func(depName string) (*model.ClassGraph, error)) error {

	projectImports := make([]string, len(fileGraph.Imports)) // all the project imports ready to go
	if breakSimbol, ok := packageBreakerSimbols[lang]; ok {
		for i, v := range fileGraph.Imports {
			splitted := strings.Split(v, breakSimbol)

			impName := splitted[len(splitted)-1] //here we trust that this is a project dependency
			projectImports[i] = impName
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
			//build the file
			n, e := cb(v) //todo: this is not working for some reason

			if e != nil {
				return e
			}

			v1.Edges = append(v1.Edges, model.GraphEdge{
				To:     n,
				From:   fileGraph,
				Weight: 0, // should be the count of times this import is used
			})
		}
	}

	return nil
}
