package ports

import (
	"raidline/ripple/core/graph/model"
)

type FileWatcherPort interface {
	Watch(dirs []string) (<-chan string, error)
}

type ProjectGraphQuerier interface {
	Exists(filename string) bool
	FindAllWithEdge(filename string) []string
}

type ProjectGraphWriter interface {
	CreateGraphForFile(
		filename string,
		fileGraph *model.ClassGraph,
		seen map[string]bool,
		fileNameToFileScan map[string]*model.FileScan,
		onFileCallback func(fileScan *model.FileScan) (*model.ClassGraph, error)) (*model.GraphVertice, error)
}
