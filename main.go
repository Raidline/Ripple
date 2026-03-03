package main

import (
	"context"
	"raidline/ripple/chat"
	"raidline/ripple/core"
	"raidline/ripple/core/events"
	"raidline/ripple/core/graph"
	"raidline/ripple/pgk/assertions"
	"raidline/ripple/pgk/logger"
	"raidline/ripple/ui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lang, watchMode := ui.NewCli().Init()

	pg := graph.CreateProjectGraph()
	fileChangesWatcher, err := events.NewWatcher()
	fileEventListener, lErr := events.NewFileListener(pg)

	if err != nil {
		panic(err)
	}

	if lErr != nil {
		panic(lErr)
	}

	s, e := core.NewService(pg, fileChangesWatcher, fileEventListener)

	assertions.NonError(e, "Could not start service")

	//todo: maybe service could return a struct that contains everything that we would need to control the TUI
	serviceSt, err := s.Orchestrate(ctx, absPath, *lang, *watchMode)

	assertions.NonError(err, "Service gave an error while Orchestrating")

	go func() {
		e := runTUI(pg, *watchMode, serviceSt)

		if e != nil {
			cancel()
		}

		cancel()
	}()

	<-ctx.Done()

	logger.Info("Shutting down background services...")
	serviceSt.Wg.Wait()
	logger.Info("Bye!")
}

func runTUI(pg graph.ProjectQuerier, watchMode bool, stateControl *core.StateControl) error {
	querier := chat.NewQuerier(pg)
	t := chat.NewTui(querier, watchMode, stateControl)

	e := t.Init()

	return e
}
