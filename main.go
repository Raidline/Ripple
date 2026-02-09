package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"raidline/ripple/chat"
	"raidline/ripple/core"
	"raidline/ripple/core/graph"
)

func main() {
	rootPath := flag.String("path", ".", "The root path of the project to analyze")
	lang := flag.String("lang", ".", "The main language of the project")

	flag.Parse()

	absPath, err := filepath.Abs(*rootPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Path [%s] does not exist.\n", absPath)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: Path [%s] is a file, but a directory is required.\n", absPath)
		os.Exit(1)
	}

	fmt.Printf("Starting Analyzer on: %s\n", absPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, e := core.NewService()

	if e != nil {
		panic(e)
	}

	serviceWg, err := s.Orchestrate(ctx, absPath, *lang)

	if err != nil {
		panic(err)
	}

	go func() {
		e := runTUI(s.ProjectGraph)

		if e != nil {
			cancel()
		}

		cancel()
	}()

	<-ctx.Done()

	fmt.Println("Shutting down background services...")
	serviceWg.Wait()
	fmt.Println("Bye!")
}

func runTUI(pg *graph.ProjectGraphAggregator) error {
	querier := chat.NewQuerier(pg)
	t := chat.NewTui(querier)

	e := t.Init()

	return e
}
