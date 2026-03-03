package application

import (
	"fmt"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
	"raidline/ripple/errors"
	"raidline/ripple/infra/file"
	"raidline/ripple/pgk/logger"
	"strings"
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
		debugProjectGraph(w.state.Graph)
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

func debugProjectGraph(graph *model.ProjectGraph) {
	if graph == nil || len(graph.Vertices) == 0 {
		logger.Debug("--- 🕸️ Project Graph is Empty ---\n")
	}

	var sb strings.Builder
	sb.WriteString("\n=== 🕸️  PROJECT GRAPH SNAPSHOT ===\n")
	fmt.Fprintf(&sb, "Total Vertices: %d\n\n", len(graph.Vertices))

	// 1. Create a reverse-lookup map to translate memory pointers back to filenames.
	// This makes the lookup O(1) instead of searching the map for every single edge.
	nameMap := make(map[*model.GraphVertice]string)
	for name, v := range graph.Vertices {
		nameMap[v] = name
	}

	// Helper function to safely get a name from a pointer
	getName := func(v *model.GraphVertice) string {
		if name, exists := nameMap[v]; exists {
			return name
		}
		return "<Unknown/Dangling Pointer>"
	}

	// 2. Iterate through every file in the graph
	for filename, vertex := range graph.Vertices {
		fmt.Fprintf(&sb, "📍 [%s]\n", filename)

		// --- OUTBOUND EDGES (Files this file depends on) ---
		sb.WriteString("   ├─ Outbound (Depends on):\n")
		if len(vertex.Edges) == 0 {
			sb.WriteString("   │  └─ (none)\n")
		} else {
			for i, edge := range vertex.Edges {
				connector := "   │  ├─▶"
				if i == len(vertex.Edges)-1 {
					connector = "   │  └─▶"
				}
				fmt.Fprintf(&sb, "%s %s\n", connector, getName(edge.To))
			}
		}

		// --- INBOUND EDGES (Files that depend on this file) ---
		sb.WriteString("   └─ Inbound (Used by):\n")
		if len(vertex.InboundEdges) == 0 {
			sb.WriteString("      └─ (none)\n\n")
		} else {
			for i, edge := range vertex.InboundEdges {
				connector := "      ├─◀"
				if i == len(vertex.InboundEdges)-1 {
					connector = "      └─◀"
				}
				// Notice we use edge.From here, because it's an inbound connection
				fmt.Fprintf(&sb, "%s %s\n", connector, getName(edge.From))
			}
			sb.WriteString("\n") // Extra newline for spacing between files
		}
	}

	sb.WriteString("==================================\n")

	logger.Debug("%s", sb.String())
}
