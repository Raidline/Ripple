package model

import (
	"fmt"
	"raidline/ripple/pgk/logger"
	"strings"
)

// todo: this should not belong here...
type LiveChangeMsg struct {
	CausingFile string
	Impacts     []string
}

// --- Graph Structures ---

type GraphEdge struct {
	To     *GraphVertice // where it is connected to
	From   *GraphVertice // where it came from
	Weight int           // this is not used at the moment - but will be used to know how much depedency there is between both
}

func CreateEdge(to *GraphVertice, from *GraphVertice) GraphEdge {
	return GraphEdge{
		To:     to,
		From:   from,
		Weight: 0, // should be the count of times this import is used
	}
}

type GraphVertice struct {
	Node         *ClassGraph
	Edges        []GraphEdge
	InboundEdges []GraphEdge
}

func CreateVertice(cg *ClassGraph) *GraphVertice {
	return &GraphVertice{
		Node:         cg,
		Edges:        make([]GraphEdge, 0),
		InboundEdges: make([]GraphEdge, 0),
	}
}

type ProjectGraph struct {
	Vertices map[string]*GraphVertice // string represent the file name as "File" without the extensions (.java/.go, etc...)
}

func (graph *ProjectGraph) Debug() {
	if len(graph.Vertices) == 0 {
		logger.Debug("--- 🕸️ Project Graph is Empty ---\n")
	}

	var sb strings.Builder
	sb.WriteString("\n=== 🕸️  PROJECT GRAPH SNAPSHOT ===\n")
	fmt.Fprintf(&sb, "Total Vertices: %d\n\n", len(graph.Vertices))

	// 1. Create a reverse-lookup map to translate memory pointers back to filenames.
	// This makes the lookup O(1) instead of searching the map for every single edge.
	nameMap := make(map[*GraphVertice]string)
	for name, v := range graph.Vertices {
		nameMap[v] = name
	}

	// Helper function to safely get a name from a pointer
	getName := func(v *GraphVertice) string {
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

type ClassGraph struct {
	ClassName string
	Fields    []Field
	Methods   []Method
	Imports   []string
}

type Field struct {
	Type string // to know if it is a dependency it needs to be in the imports
	Name string
}

type Method struct {
	Name       string
	ReturnType string
	Params     []Param
	Calls      []MethodCall // "Edges" of your graph
}

type Param struct {
	Type string
	Name string
}

type MethodCall struct {
	Target string // e.g. "this.repo" or "service"
	Method string // e.g. "persist"
}
