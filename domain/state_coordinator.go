package domain

import (
	"context"
	"raidline/ripple/domain/model"
	"sync"
)

// the Graph this might need a mutex around if we write and read at the same time (not needed at the moment because we only write at build time)

type StateCoordinator struct {
	Ctx            context.Context
	LiveChangeChan <-chan model.LiveChangeMsg
	Graph          *model.ProjectGraph //todo: make private and have methods to access?
	wg             *sync.WaitGroup     //global waitGroup for top-level goroutines
}

//this should deal with all creation of goroutine and such

// todo: receive the rest (or set)
func NewStateCoordinator(ctx context.Context) *StateCoordinator {
	return &StateCoordinator{
		Ctx:            ctx,
		wg:             &sync.WaitGroup{},
		LiveChangeChan: nil,
	}
}

func (s *StateCoordinator) CreateGoroutine(f func()) {
	s.wg.Go(f)
}

func (s *StateCoordinator) Wait() {
	s.wg.Wait()
}

func (s *StateCoordinator) ResetGraph() {
	s.Graph = &model.ProjectGraph{
		Vertices: map[string]*model.GraphVertice{},
	}
}

func (s *StateCoordinator) IsLiveWatchMode() bool {
	return s.LiveChangeChan != nil
}
