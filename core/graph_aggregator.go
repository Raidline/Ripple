package core

import (
	"raidline/ripple/core/languages"
	"raidline/ripple/core/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
)

type ProjectGraphAggregator struct {
	graph *model.ProjectGraph
}

func Create() *ProjectGraphAggregator {
	return &ProjectGraphAggregator{}
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

		appErr := appendToCurrentGraph(fileGraph)

		if appErr != nil {
			return appErr
		}
	}

	return nil
}

func appendToCurrentGraph(fileGraph *model.ClassGraph) error {
	panic("unimplemented")
}
