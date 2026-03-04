package services

import (
	"raidline/ripple/core/graph/model"
	"raidline/ripple/domain"
	"raidline/ripple/domain/ports"
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
		v := model.CreateVertice(fileGraph)
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

			edge := model.CreateEdge(vertice, vert)

			vert.Edges = append(vert.Edges, edge)
			vertice.InboundEdges = append(vertice.InboundEdges, edge)
		} else {
			//build the file
			newVertice, e := cb(scan) //todo: if we see this causes to much memory or time, we can skip this node here and implement the logic where we see if this node is already connected to any another node and make that connection
			// on a large project probably this make more sense, because you will end up with the Stacj going crazy

			if e != nil {
				return e
			}

			edge := model.CreateEdge(newVertice, vert)

			vert.Edges = append(vert.Edges, edge)
			newVertice.InboundEdges = append(newVertice.InboundEdges, edge)
		}
	}

	return nil
}
