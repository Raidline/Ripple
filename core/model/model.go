package model

// --- Graph Structures ---

type GraphEdge struct {
	To     *ClassGraph
	Weight int // this is not used at the moment - but will be used to know how much depedency there is between both
}

type ProjectGraph struct {
	Graph [][]GraphEdge
}

type ClassGraph struct {
	ClassName string
	Fields    []Field
	Methods   []Method
	Imports   []string
}

type Field struct {
	Type string
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
