package main

import (
	"context"
	"raidline/ripple/chat"
	"raidline/ripple/domain"
	"raidline/ripple/pgk/logger"
	"raidline/ripple/ui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stateCoordinator := domain.NewStateCoordinator(ctx)

	lang, watchMode := ui.NewCli().Init()

	go func() {
		e := runTUI(pg, *watchMode, serviceSt)

		if e != nil {
			cancel()
		}

		cancel()
	}()

	<-ctx.Done()

	logger.Info("Shutting down background services...")
	stateCoordinator.Wait()
	logger.Info("Bye!")
}

func runTUI(pg graph.ProjectQuerier, watchMode bool, stateControl *core.StateControl) error {
	querier := chat.NewQuerier(pg)
	t := chat.NewTui(querier, watchMode, stateControl)

	e := t.Init()

	return e
}
