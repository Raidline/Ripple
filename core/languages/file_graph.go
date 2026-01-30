package languages

import (
	"context"
	"raidline/ripple/pgk"

	sitter "github.com/smacker/go-tree-sitter"
)

// --- Graph Structures ---

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

func BuildFileGraph(file *pgk.FileScan, analyser LanguageAnalyser) (*ClassGraph, error) {
	source, fileErr := convertFileScan(file)

	if fileErr != nil {
		return nil, fileErr
	}

	parser := sitter.NewParser()
	lang := analyser.GetLanguage()
	parser.SetLanguage(lang)
	graph := &ClassGraph{}

	tree, err := parser.ParseCtx(context.Background(), nil, source)

	if err != nil {
		return nil, err
	}

	root := tree.RootNode()

	execQuery(analyser.GetStructQuery(), root, lang, source, func(tag, content string, n *sitter.Node) {
		if tag == "class.name" {
			graph.ClassName = content
		}

		if tag == "import" {
			graph.Imports = append(graph.Imports, content)
		}
	})

	execQuery(analyser.GetFieldQuery(), root, lang, source, func(tag, content string, n *sitter.Node) {
		curField := analyser.MapField(tag, content)
		// If the field name is set, we assume the field is "complete"
		if curField.Name != "" {
			graph.Fields = append(graph.Fields, curField)
			curField = Field{}
		}
	})

	// 3. Methods
	mQuery, _ := sitter.NewQuery([]byte(analyser.GetMethodQuery()), analyser.GetLanguage())
	qc := sitter.NewQueryCursor()
	qc.Exec(mQuery, root)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		mNode := match.Captures[0].Node
		method := Method{}

		// Map the method metadata (Name, ReturnType) using the analyzer
		// This processes the tags like @method.name and @method.return
		for _, capture := range match.Captures {
			tag := mQuery.CaptureNameForId(capture.Index)
			content := capture.Node.Content(source)
			method = analyser.MapMethod(tag, content)
		}

		// Internal: Params
		execQuery(analyser.GetParamQuery(), mNode, lang, source, func(tag, content string, n *sitter.Node) {

			curParam := analyser.MapParam(tag, content)
			if curParam.Name != "" {
				method.Params = append(method.Params, curParam)
				curParam = Param{}
			}
		})

		// Internal: Calls
		execQuery(analyser.GetCallQuery(), mNode, lang, source, func(tag, content string, n *sitter.Node) {
			curCall := analyser.MapCall(tag, content)
			if curCall.Method != "" {
				method.Calls = append(method.Calls, curCall)
				curCall = MethodCall{}
			}
		})

		graph.Methods = append(graph.Methods, method)
	}
	return graph, nil
}

func convertFileScan(file *pgk.FileScan) ([]byte, error) {
	if file.Err != nil {
		return nil, file.Err
	}

	var source []byte

	for line := range file.Lines {
		source = append(source, []byte(line)...)
	}

	return source, nil
}

func execQuery(qStr string, node *sitter.Node,
	lang *sitter.Language,
	source []byte, cb func(tag, content string, n *sitter.Node)) {
	q, err := sitter.NewQuery([]byte(qStr), lang)
	if err != nil {
		return
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, node)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, capture := range match.Captures {
			tag := q.CaptureNameForId(capture.Index)
			content := capture.Node.Content(source)
			cb(tag, content, capture.Node)
		}
	}
}
