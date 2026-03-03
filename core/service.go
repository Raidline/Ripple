package core

import (
	"context"
	"log"
	"raidline/ripple/core/events"
	"raidline/ripple/core/graph"
	"raidline/ripple/core/graph/creeper"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk/logger"
	"sync"
)

// todo: this should be in another folder and every goroutine should be created there to centralize the state
// this should not be here, and the service should receive this and not return this
type StateControl struct {
	Ctx        context.Context
	Wg         *sync.WaitGroup
	WatcherRes chan model.LiveChangeMsg
}

type Service struct {
	aggregator graph.ProjectGraphWriter
	watcher    *events.FileWatcher
	listener   *events.FileEventListener
}

func NewService(pg graph.ProjectGraphWriter, watcher *events.FileWatcher, listener *events.FileEventListener) (*Service, error) { // should receive params from input , tbd
	return &Service{
		aggregator: pg,
		watcher:    watcher,
		listener:   listener,
	}, nil
}

func (s *Service) Orchestrate(ctx context.Context,
	root string, lang string, watchMode bool) (*StateControl, error) {

	//todo: in here we need to see if we already have the file serving as DB for the graph.
	// if that is the case get the graph from there and run the below logic in a goroutine and update the graph in the background

	//maintain the return of the WaitGroup,
	// as per the improvements part we need to run in a goroutine the creeper to catch new changes that happened while the tool was not running
	st := &StateControl{}
	wg := &sync.WaitGroup{}
	st.Ctx = ctx
	st.Wg = wg

	log.Printf("creeping project in dir : [%s] for lang : [%s]", root, lang)
	creepRes, err := creeper.CreepDir(root)

	if err != nil {
		return nil, err
	}

	if watchMode {
		st.WatcherRes = make(chan model.LiveChangeMsg, 10)
		logger.Debug("Watch mode enabled, listening for changes...")
		eventChan, wErr := s.watcher.Watch(ctx, creepRes.Dirs)

		if wErr != nil {
			panic(wErr)
		}

		wg.Go(func() {
			s.listener.Listen(ctx, eventChan, st.WatcherRes)
		})
	}

	//todo: nexte step - service for aggregation (not use case because it does not need other dependencies)

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
	gErr := s.aggregator.Aggregate(creepRes.Files, wantedLang)

	if gErr != nil {
		return nil, gErr
	}

	return st, nil
}
