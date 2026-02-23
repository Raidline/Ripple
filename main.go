package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"raidline/ripple/chat"
	"raidline/ripple/core"
	"raidline/ripple/core/events"
	"raidline/ripple/core/graph"
	"raidline/ripple/pgk/assertions"
	"raidline/ripple/pgk/logger"
)

func main() {
	rootPath := flag.String("path", ".", "The root path of the project to analyze")
	lang := flag.String("lang", ".", "The main language of the project")
	watchMode := flag.Bool("watch", false, "Watch live changes to the project")
	debugMode := flag.Bool("debug", false, "Turn on debug mode")

	flag.Parse()

	absPath, err := filepath.Abs(*rootPath)
	assertions.NonError(err, fmt.Sprintf("Error resolving path: %v\n", err))

	info, err := os.Stat(absPath)
	assertions.Condition(!os.IsNotExist(err), fmt.Sprintf("Error: Path [%s] does not exist.\n", absPath))
	assertions.Condition(info.IsDir(), fmt.Sprintf("Error: Path [%s] is a file, but a directory is required.\n", absPath))

	logger.Init(*debugMode)

	logger.Info("Starting Analyzer on: %s", absPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		e := runTUI(pg, *watchMode, serviceSt.WatcherRes)

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

func runTUI(pg graph.ProjectQuerier, watchMode bool, fileChangesChan <-chan []string) error {
	querier := chat.NewQuerier(pg)
	t := chat.NewTui(querier, watchMode, fileChangesChan)

	e := t.Init()

	return e
}
