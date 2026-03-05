package application

import (
	"raidline/ripple/domain"
	"raidline/ripple/domain/model"
	"raidline/ripple/domain/ports"
	"raidline/ripple/pgk/logger"
)

type WatchFileChangesUseCase struct {
	watcher ports.FileWatcherPort
	querier ports.ProjectGraphQuerier
	state   *domain.StateCoordinator
}

func NewWatchFileUseCase(watcher ports.FileWatcherPort,
	querier ports.ProjectGraphQuerier, state *domain.StateCoordinator) *WatchFileChangesUseCase {
	return &WatchFileChangesUseCase{
		watcher: watcher,
		querier: querier,
		state:   state,
	}
}

func (w *WatchFileChangesUseCase) WatchFileChange(dirs []string) error {
	logger.Debug("Watch mode enabled, listening for changes...")
	liveChan := make(chan model.LiveChangeMsg, 10)

	eventChan, wErr := w.watcher.Watch(dirs)

	if wErr != nil {
		return wErr
	}

	w.state.CreateGoroutine(func() {

		for {
			select {
			case <-w.state.Ctx.Done():
				logger.Debug("Context cancelled, cleaning up...")
				return
			case file, ok := <-eventChan:
				if !ok {
					logger.Debug("Consumer: Event channel closed, exiting...")
					return
				}

				if w.querier.Exists(file) {
					impacts := w.querier.FindAllWithEdge(file)
					liveChan <- model.LiveChangeMsg{
						CausingFile: file,
						Impacts:     impacts,
					}
				} else {
					logger.Debug("The file we received update from is not in the graph. \n")
				}
			}
		}
	})

	w.state.LiveChangeChan = liveChan

	return nil
}
