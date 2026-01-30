package model

// --- Graph Structures ---

type GraphEdge struct {
	To     *ClassGraph // where it is connected to
	From   *ClassGraph // where it came from
	Weight int         // this is not used at the moment - but will be used to know how much depedency there is between both
}

type GraphVertice struct {
	Node  *ClassGraph
	Edges []GraphEdge
}

type ProjectGraph struct {
	Vertices map[string]GraphVertice
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
