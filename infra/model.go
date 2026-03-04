package infra

import (
	"iter"
)

type CreepScanResult struct {
	Dirs  []string
	Files []*FileScan
}

type FileScan struct {
	Dir   string
	Name  string
	Lines iter.Seq[string]
}
