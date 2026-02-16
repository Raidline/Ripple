package graph

import (
	"fmt"
	"raidline/ripple/core/graph/file"
	"raidline/ripple/core/graph/languages"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/errors"
	"raidline/ripple/pgk/logger"
	"raidline/ripple/pgk/structures"
	"slices"
	"strings"
)

type ProjectQuerier interface {
	Exists(filename string) bool
	FindAllWithEdge(filename string) []string
}

type ProjectGraphWriter interface {
	Aggregate(files []*model.FileScan, wantedLang languages.Language) error
}

type ProjectGraph struct {
	graph *model.ProjectGraph
}

func CreateProjectGraph() *ProjectGraph {
	return &ProjectGraph{
		graph: &model.ProjectGraph{
			Vertices: make(map[string]*model.GraphVertice, 0),
		},
	}
}

// --- Graph read operations --- \\
func (pg *ProjectGraph) FindAllWithEdge(filename string) []string {
	if len(pg.graph.Vertices) == 0 {
		return make([]string, 0)
	}
	source := pg.graph.Vertices[filename]

	q := structures.NewQueue[*model.GraphVertice]()
	seen := make(map[string]bool, len(pg.graph.Vertices))
	prev := make(map[string]*model.GraphVertice, len(pg.graph.Vertices))
	result := make([]string, 0)

	for i := range prev {
		prev[i] = nil
	}

	seen[source.Node.ClassName] = true
	q.Enqueue(source)

	for q.Length > 0 {

		curr, err := q.Deque()

		if err != nil { // no more items
			break
		}

		seen[curr.Node.ClassName] = true

		for _, v := range pg.graph.Vertices[curr.Node.ClassName].Edges {

			if seen[v.From.Node.ClassName] {
				continue
			}

			seen[v.From.Node.ClassName] = true
			prev[v.From.Node.ClassName] = curr

			q.Enqueue(v.From)
		}
	}

	curr := pg.graph.Vertices[filename]
	out := make([]string, 0)

	for prev[curr.Node.ClassName] != nil {
		out = append(out, curr.Node.ClassName)
		curr = prev[curr.Node.ClassName]
	}

	if len(out) == 0 {
		return []string{}
	}

	out = append(out, source.Node.ClassName)
	slices.Reverse(out)

	return result
}

func (pg *ProjectGraph) Exists(filename string) bool {
	_, ok := pg.graph.Vertices[filename]

	return ok
}

// --- Graph write operations --- \\
func (agg *ProjectGraph) Aggregate(files []*model.FileScan, wantedLang languages.Language) error {

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

	agg.createProjectGraph(fileNameToFileScan, fileAnalyser, wantedLang)

	go func() {
		debugProjectGraph(agg.graph)
	}()

	return nil
}

func (agg *ProjectGraph) createProjectGraph(
	fileNameToFileScan map[string]*model.FileScan,
	fileAnalyser languages.LanguageAnalyser,
	wantedLang languages.Language) error {

	seen := make(map[string]bool, 0)

	for _, file := range fileNameToFileScan {
		_, e := agg.createGraphForFile(file, seen, fileNameToFileScan, fileAnalyser, wantedLang)

		if e != nil {
			return e
		}
	}

	return nil
}

func (agg *ProjectGraph) createGraphForFile(fileScan *model.FileScan,
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

	if vertice, ok := agg.graph.Vertices[fileGraph.ClassName]; ok {
		// this might already exist but as a from dependency, now we need to add the to's
		agg.connectEdgesToVertice(vertice, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			return agg.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return vertice, nil
	} else {
		v := &model.GraphVertice{
			Node:  fileGraph,
			Edges: make([]model.GraphEdge, 0),
		}
		// todo(the fields and method info to get the weight of each import)
		agg.graph.Vertices[fileGraph.ClassName] = v

		//Edges will be added by memory reference
		agg.connectEdgesToVertice(v, fileNameToFileScan, fileGraph, func(fileScan *model.FileScan) (*model.GraphVertice, error) {
			return agg.createGraphForFile(fileScan, seen, fileNameToFileScan, fileAnalyser, wantedLang)
		})

		seen[fileScan.Name] = true

		return v, nil
	}
}

func (agg *ProjectGraph) connectEdgesToVertice(vert *model.GraphVertice,
	fileNameToFileScan map[string]*model.FileScan,
	fileGraph *model.ClassGraph, cb func(fileScan *model.FileScan) (*model.GraphVertice, error)) error {

	fieldToDependency := make(map[string]*model.FileScan, 0)

	for _, f := range fileGraph.Fields {
		if v, ok := fileNameToFileScan[f.Type]; ok {
			fieldToDependency[f.Type] = v
		}
	}

	for name, scan := range fieldToDependency {

		if vertice, ok := agg.graph.Vertices[name]; ok {
			vert.Edges = append(vert.Edges, model.GraphEdge{
				To:     vertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			})
		} else {
			//build the file
			newVertice, e := cb(scan) //todo: if we see this causes to much memory or time, we can skip this node here and implement the logic where we see if this node is already connected to any another node and make that connection
			// on a large project probably this make more sense, because you will end up with the Stacj going crazy

			if e != nil {
				return e
			}

			vert.Edges = append(vert.Edges, model.GraphEdge{
				To:     newVertice,
				From:   vert,
				Weight: 0, // should be the count of times this import is used
			})
		}
	}

	return nil
}

func debugProjectGraph(graph *model.ProjectGraph) {
	var sb strings.Builder

	sb.WriteString("\n--- Project Graph Snapshot ---\n")

	if graph == nil || len(graph.Vertices) == 0 {
		sb.WriteString("Graph is empty.\n")
		logger.Debug(sb.String())
		return
	}

	nameMap := make(map[*model.GraphVertice]string)
	for name, v := range graph.Vertices {
		nameMap[v] = name
	}

	sb.WriteString(fmt.Sprintf("Total Vertices: %d\n", len(graph.Vertices)))

	for filename, vertex := range graph.Vertices {
		sb.WriteString(fmt.Sprintf("%s\n", filename))

		if len(vertex.Edges) == 0 {
			sb.WriteString("  └── (no dependencies)\n")
			continue
		}

		for i, edge := range vertex.Edges {
			connector := "  ├──"
			if i == len(vertex.Edges)-1 {
				connector = "  └──"
			}

			targetName := nameMap[edge.To]
			if targetName == "" {
				targetName = "External/Unknown"
			}

			sb.WriteString(fmt.Sprintf("%s %s (w:%d)\n", connector, targetName, edge.Weight))
		}
	}

	sb.WriteString("--- End of Snapshot ---\n")

	logger.Debug(sb.String())
}
