package ports

import (
	"raidline/ripple/domain/model"
	"raidline/ripple/infra"
)

type DirectoryCreeperPort interface {
	CreepDir(dir string) (*infra.CreepScanResult, error)
}

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
		fileNameToFileScan map[string]*infra.FileScan,
		onFileCallback func(fileScan *infra.FileScan) (*model.ClassGraph, error)) (*model.GraphVertice, error)
}
