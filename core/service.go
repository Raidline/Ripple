package core

import (
	"raidline/ripple/core/graph"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/watcher"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
)

type Service struct {
	projectGraph *graph.ProjectGraphAggregator
	watcher      *watcher.FileWatcher
}

func NewService() (*Service, error) { // should receive params from input , tbd
	graphAggregator := graph.Create()
	fileWatcher, err := watcher.NewWatcher(graphAggregator)

	if err != nil {
		return nil, err
	}

	return &Service{
		projectGraph: graphAggregator,
		watcher:      fileWatcher,
	}, nil
}

func (s *Service) Orchestrate(root string, lang string) error {

	creepRes, err := pgk.CreepDir(root)

	if err != nil {
		return err
	}

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
	// in a optimal world we would get the lang by creeping the project.tbd
	gErr := s.projectGraph.Aggregate(creepRes.Files, wantedLang)

	if gErr != nil {
		return gErr
	}

	wErr := s.watcher.Watch(creepRes.Dirs)

	if wErr != nil {
		return wErr
	}

	//todo: make first the query to the graph to make sure we can get what we want
	//todo: only then do the watcher

	return nil
}
