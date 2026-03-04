package application

import (
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
	"raidline/ripple/errors"
	"raidline/ripple/infra/file"
)

type GraphBuildUseCase struct {
	state       *domain.StateCoordinator
	graphWriter ports.ProjectGraphWriter
	repo        *file.FileGraphRepo
}

func (w *GraphBuildUseCase) Build(lang string) error {
	var languageErr error
	var wantedLang languages.Language

	if lang == string(languages.JAVA) {
		wantedLang = languages.JAVA
	} else if lang == string(languages.TS) {
		wantedLang = languages.TS
	} else {
		languageErr = errors.NewLanguageNotSupportedError(lang)
	}

	if languageErr != nil {
		return languageErr
	}

	files := w.state.DirCreepResult.Files

	//todo: in a optimal world we would get the lang by creeping the project.
	if files == nil {
		return errors.NewEmptySequenceError("files sequence")
	}

	fileAnalyser, e := languages.GetAnalyser(wantedLang)

	if e != nil {
		return e
	}

	fileNameToFileScan := make(map[string]*model.FileScan, 0)

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
	fileNameToFileScan map[string]*model.FileScan,
	fileAnalyser languages.LanguageAnalyser) error {

	var (
		fileCallback = func(fileScan *model.FileScan) (*model.ClassGraph, error) {
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
