package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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

	//todo: here we can launch some kind of TUI, or interactive "chat" where user can ask questions about the project
	go func() {
		runTUI(s.ProjectGraph)

		cancel()
	}()

	<-ctx.Done()

	fmt.Println("Shutting down background services...")
	serviceWg.Wait()
	fmt.Println("Bye!")
}

func runTUI(pg *graph.ProjectGraphAggregator) {
	// we need to be aware that in the case that we write in the graph (new file for example)
	// if we try to access the graph node in here while it is being written we will have a race condition
	// when do this also create a mutex in the model ClassGraph to safe-guard that
	fmt.Println("I AM A TUIIIII")
}
