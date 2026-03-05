package main

import (
	"context"
	"raidline/ripple/application"
	"raidline/ripple/domain"
	"raidline/ripple/domain/services"
	"raidline/ripple/infra/directories"
	"raidline/ripple/infra/file"
	"raidline/ripple/pgk/logger"
	"raidline/ripple/ui"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stateCoordinator := domain.NewStateCoordinator(ctx)

	watcher, e := services.NewWatcher(stateCoordinator)
	if e != nil {
		panic(e)
	}

	querier := services.NewQuerier(stateCoordinator)
	graphWriter := services.NewGraphWriter(stateCoordinator)
	directoryCreeperPort := directories.NewCreeper()

	watchFileUseCase := application.NewWatchFileUseCase(watcher, querier, stateCoordinator)
	graphBuildUseCase := application.NewGraphBuildUseCase(stateCoordinator, graphWriter, file.NewFileGraphRepo(),
		directoryCreeperPort)

	ui.NewCli(watchFileUseCase, graphBuildUseCase, directoryCreeperPort).Init()

	go func() {
		t := ui.NewTui(stateCoordinator)

		e := t.Init()

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
