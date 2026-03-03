package services

import (
	"fmt"
	"raidline/ripple/core/graph/file"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
	"raidline/ripple/errors"
	"raidline/ripple/pgk/logger"
	"strings"
)

type graphWriter struct {
	state *domain.StateCoordinator
}

func NewGraphWriter(state *domain.StateCoordinator) ports.ProjectGraphWriter {
	return &graphWriter{
		state: state,
	}
}

func (w *graphWriter) CreateGraphForFile(
	filename string,
	fileGraph *model.ClassGraph,
	seen map[string]bool,
	fileNameToFileScan map[string]*model.FileScan,
	onFileCallback func(fileScan *model.FileScan) (*model.ClassGraph, error)) (*model.GraphVertice, error) {

	if vertice, ok := w.state.Graph.Vertices[fileGraph.ClassName]; ok {
		// this might already exist but as a from dependency, now we need to add the to's
		w.connectEdgesToVertice(vertice, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			fileGraph, err := onFileCallback(fileScan)

			if err != nil {
				return nil, err
			}

			return w.CreateGraphForFile(filename, fileGraph, seen, fileNameToFileScan, onFileCallback)
		})

		seen[filename] = true

		return vertice, nil
	} else {
		v := &model.GraphVertice{
			Node:  fileGraph,
			Edges: make([]model.GraphEdge, 0),
		}
		// todo(the fields and method info to get the weight of each import)
		w.state.Graph.Vertices[fileGraph.ClassName] = v

		//Edges will be added by memory reference
		w.connectEdgesToVertice(v, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			fileGraph, err := onFileCallback(fileScan)

			if err != nil {
				return nil, err
			}

			return w.CreateGraphForFile(filename, fileGraph, seen, fileNameToFileScan, onFileCallback)
		})

		seen[filename] = true

		return v, nil
	}
}

func (w *graphWriter) Build(lang string) error {
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

	w.createProjectGraph(fileNameToFileScan, fileAnalyser, wantedLang)

	go func() {
		debugProjectGraph(w.state.Graph)
	}()

	return nil
}

func (w *graphWriter) createProjectGraph(
	fileNameToFileScan map[string]*model.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) error {

	seen := make(map[string]bool, 0)

	for _, file := range fileNameToFileScan {
		_, e := w.createGraphForFile(file, seen, fileNameToFileScan, fileAnalyser, wantedLang)

		if e != nil {
			return e
		}
	}

	return nil
}

func (w *graphWriter) createGraphForFile(fileScan *model.FileScan,
	seen map[string]bool,
	fileNameToFileScan map[string]*model.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) (*model.GraphVertice, error) {

	if seen[fileScan.Name] {
		return nil, nil
	}

	if fileScan == nil {
		panic("Scan of the file cannot be empty")
	}

	fileGraph, fileGErr := file.BuildFileGraph(fileScan, fileAnalyser)

	if fileGErr != nil {
		return nil, fileGErr
	}

	if vertice, ok := w.state.Graph.Vertices[fileGraph.ClassName]; ok {
		// this might already exist but as a from dependency, now we need to add the to's
		w.connectEdgesToVertice(vertice, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			return w.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return vertice, nil
	} else {
		v := &model.GraphVertice{
			Node:  fileGraph,
			Edges: make([]model.GraphEdge, 0),
		}
		// todo(the fields and method info to get the weight of each import)
		w.state.Graph.Vertices[fileGraph.ClassName] = v

		//Edges will be added by memory reference
		w.connectEdgesToVertice(v, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			return w.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return v, nil
	}
}

func (w *graphWriter) connectEdgesToVertice(vert *model.GraphVertice,
	fileNameToFileScan map[string]*model.FileScan,
	fileGraph *model.ClassGraph, cb func(fileScan *model.FileScan) (*model.GraphVertice, error)) error {

	fieldToDependency := make(map[string]*model.FileScan, 0)

	for _, f := range fileGraph.Fields {
		if v, ok := fileNameToFileScan[f.Type]; ok {
			fieldToDependency[f.Type] = v
		}
	}

	for name, scan := range fieldToDependency {

		if vertice, ok := w.state.Graph.Vertices[name]; ok {

			edge := model.GraphEdge{
				To:     vertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			}

			vert.Edges = append(vert.Edges, edge)
			vertice.InboundEdges = append(vertice.InboundEdges, edge)
		} else {
			//build the file
			newVertice, e := cb(scan) //todo: if we see this causes to much memory or time, we can skip this node here and implement the logic where we see if this node is already connected to any another node and make that connection
			// on a large project probably this make more sense, because you will end up with the Stacj going crazy

			if e != nil {
				return e
			}

			edge := model.GraphEdge{
				To:     newVertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			}

			vert.Edges = append(vert.Edges, edge)
			newVertice.InboundEdges = append(newVertice.InboundEdges, edge)
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
