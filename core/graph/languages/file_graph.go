package languages

import (
	"context"
	"raidline/ripple/core/graph/model"
	"raidline/ripple/pgk"

	sitter "github.com/smacker/go-tree-sitter"
)

func BuildFileGraph(file *pgk.FileScan, analyser LanguageAnalyser) (*model.ClassGraph, error) {
	source := convertFileScan(file)

	parser := sitter.NewParser()
	parser.SetLanguage(analyser.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, err
	}
	root := tree.RootNode()
	graph := &model.ClassGraph{}

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	runQuery(analyser.GetStructQuery(), root, source, func(tag, content string, n *sitter.Node) {
		switch tag {
		case "class.name":
			graph.ClassName = content
		case "import":
			if analyser.IsProjectImport(content) {
				graph.Imports = append(graph.Imports, content)
			}
		}
	})

	curField := model.Field{}

	runQuery(analyser.GetFieldQuery(), root, source, func(tag, content string, n *sitter.Node) {
		analyser.MapField(tag, content, &curField)
		// If the field name is set, we assume the field is "complete"
		if curField.Name != "" {
			graph.Fields = append(graph.Fields, curField)
			curField.Name = ""
		}
	})

	mQuery := analyser.GetMethodQuery()
	qc.Exec(mQuery, root)

	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		mNode := match.Captures[0].Node
		method := model.Method{}

		// Map the method metadata (Name, ReturnType) using the analyzer
		for _, capture := range match.Captures {
			tag := mQuery.CaptureNameForId(capture.Index)
			content := string(source[capture.Node.StartByte():capture.Node.EndByte()])
			analyser.MapMethod(tag, content, &method)
		}

		if mNode != nil {
			// Internal: Params
			curParam := model.Param{}
			curCall := model.MethodCall{}
			runQuery(analyser.GetParamQuery(), mNode, source, func(tag, content string, n *sitter.Node) {

				analyser.MapParam(tag, content, &curParam)
				if curParam.Name != "" {
					method.Params = append(method.Params, curParam)
					curParam.Name = ""
				}
			})

			// Internal: Calls
			runQuery(analyser.GetCallQuery(), mNode, source, func(tag, content string, n *sitter.Node) {
				analyser.MapCall(tag, content, &curCall)
				if curCall.Method != "" {
					method.Calls = append(method.Calls, curCall)
					curCall.Method = ""
				}
			})
		}

		graph.Methods = append(graph.Methods, method)
	}
	return graph, nil
}

func convertFileScan(file *pgk.FileScan) []byte {
	var source []byte

	for line := range file.Lines {
		source = append(source, []byte(line)...)
	}

	return source
}

func runQuery(q *sitter.Query, node *sitter.Node, source []byte, cb func(tag, content string, n *sitter.Node)) {
	if q == nil {
		return
	}
	qc := sitter.NewQueryCursor()
	defer qc.Close()
	qc.Exec(q, node)
	for {
		match, ok := qc.NextMatch()
		if !ok {
			break
		}
		for _, cap := range match.Captures {
			tag := q.CaptureNameForId(cap.Index)
			content := string(source[cap.Node.StartByte():cap.Node.EndByte()])
			cb(tag, content, cap.Node)
		}
	}
}
