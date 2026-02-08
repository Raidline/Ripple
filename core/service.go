package core

import (
	"context"
	"os"
	"os/signal"
	"raidline/ripple/core/events"
	"raidline/ripple/core/graph"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/errors"
	"raidline/ripple/pgk"
)

type Service struct {
	projectGraph *graph.ProjectGraphAggregator
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
		projectGraph: graphAggregator,
		watcher:      fileWatcher,
		listener:     listener,
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

	//probably need a wait group in main to wait for all the goroutines to finish

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop() // Clean up resources when main exits

	eventChan, wErr := s.watcher.Watch(ctx, creepRes.Dirs) //todo

	if wErr != nil {
		return wErr
	}

	s.listener.Listen(ctx, eventChan) //todo

	return nil
}
