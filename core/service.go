package core

import (
	"context"
	"raidline/ripple/core/events"
	"raidline/ripple/core/graph"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
	"sync"
)

type Service struct {
	ProjectGraph *graph.ProjectGraphAggregator
	watcher      *events.FileWatcher
	listener     *events.FileEventListener
}

func NewService() (*Service, error) { // should receive params from input , tbd
	graphAggregator := graph.Create()
	fileWatcher, err := events.NewWatcher(graphAggregator)
	listener, lErr := events.NewFileListener(graphAggregator)

	if err != nil {
		return nil, err
	}

	if lErr != nil {
		return nil, lErr
	}

	return &Service{
		ProjectGraph: graphAggregator,
		watcher:      fileWatcher,
		listener:     listener,
	}, nil
}

func (s *Service) Orchestrate(ctx context.Context, root string, lang string) (*sync.WaitGroup, error) {
	creepRes, err := pgk.CreepDir(root)

	if err != nil {
		return nil, err
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
		return nil, languageErr
	}

	//todo: in a optimal world we would get the lang by creeping the project.
	gErr := s.ProjectGraph.Aggregate(creepRes.Files, wantedLang)

	if gErr != nil {
		return nil, gErr
	}

	wg := &sync.WaitGroup{}

	eventChan, wErr := s.watcher.Watch(ctx, creepRes.Dirs)

	if wErr != nil {
		return nil, wErr
	}

	wg.Go(func() {
		s.listener.Listen(ctx, eventChan)
	})

	return wg, nil
}
