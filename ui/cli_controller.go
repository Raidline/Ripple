package ui

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"raidline/ripple/pgk/assertions"
	"raidline/ripple/pgk/logger"
)

type Cli struct {
}

func NewCli() *Cli {
	return &Cli{}
}

func (c *Cli) Init() (string, bool) {
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

	//todo: if watchMode launch use case

	return *lang, *watchMode
}
