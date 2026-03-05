package application

import (
	"raidline/ripple/domain"
	"raidline/ripple/domain/model"
	"raidline/ripple/domain/ports"
	"raidline/ripple/errors"
	"raidline/ripple/infra"
	"raidline/ripple/infra/file"
)

type GraphBuildUseCase struct {
	state           *domain.StateCoordinator
	graphWriter     ports.ProjectGraphWriter
	repo            *file.FileGraphRepo
	directoriesRepo ports.DirectoryCreeperPort
}

func NewGraphBuildUseCase(state *domain.StateCoordinator, graphWriter ports.ProjectGraphWriter,
	repo *file.FileGraphRepo, dirsRepo ports.DirectoryCreeperPort) *GraphBuildUseCase {
	return &GraphBuildUseCase{
		state:           state,
		graphWriter:     graphWriter,
		repo:            repo,
		directoriesRepo: dirsRepo,
	}
}

func (w *GraphBuildUseCase) Build(lang string, files []*infra.FileScan) error {
	var languageErr error
	var wantedLang model.Language

	if lang == string(model.JAVA) {
		wantedLang = model.JAVA
	} else if lang == string(model.TS) {
		wantedLang = model.TS
	} else {
		languageErr = errors.NewLanguageNotSupportedError(lang)
	}

	if languageErr != nil {
		return languageErr
	}

	//todo: in a optimal world we would get the lang by creeping the project.
	if files == nil {
		return errors.NewEmptySequenceError("files sequence")
	}

	fileAnalyser, e := model.GetAnalyser(wantedLang)

	if e != nil {
		return e
	}

	fileNameToFileScan := make(map[string]*infra.FileScan, 0)

	for _, file := range files {
		fileNameToFileScan[file.Name] = file
	}

	w.buildGraph(fileNameToFileScan, fileAnalyser)

	go func() {
		w.state.Graph.Debug()
	}()

	return nil
}

func (w *GraphBuildUseCase) buildGraph(
	fileNameToFileScan map[string]*infra.FileScan,
	fileAnalyser model.LanguageAnalyser) error {

	var (
		fileCallback = func(fileScan *infra.FileScan) (*model.ClassGraph, error) {
			fileGraph, fileGErr := w.repo.BuildFileGraph(fileScan, fileAnalyser)

			if fileGErr != nil {
				return nil, fileGErr
			}

			return fileGraph, nil
		}
	)

	seen := make(map[string]bool, 0)

	for _, fileScan := range fileNameToFileScan {

		if seen[fileScan.Name] {
			continue
		}

		if fileScan == nil {
			panic("Scan of the file cannot be empty")
		}

		fileGraph, fileGErr := w.repo.BuildFileGraph(fileScan, fileAnalyser)

		if fileGErr != nil {
			w.state.ResetGraph()

			return fileGErr
		}

		seen[fileScan.Name] = true

		_, e := w.graphWriter.CreateGraphForFile(
			fileScan.Name,
			fileGraph,
			seen,
			fileNameToFileScan,
			fileCallback)

		if e != nil {
			w.state.ResetGraph()

			return e
		}
	}

	return nil
}
