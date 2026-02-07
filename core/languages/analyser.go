package languages

import (
	"fmt"
	"raidline/ripple/core/model"
	"raidline/ripple/errors"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	java "github.com/smacker/go-tree-sitter/java"
	typescript "github.com/smacker/go-tree-sitter/typescript/typescript" // This is for .ts files
)

type Language string

const (
	JAVA Language = "java"
	TS   Language = "typescript"
)

// Generic Interface that should be implemented in order to analyse a file in a language
type LanguageAnalyser interface {
	GetLanguage() *sitter.Language
	// Queries
	GetStructQuery() *sitter.Query // Query to find the struct of the file (class name & imports)
	GetFieldQuery() *sitter.Query  // Query to find the fields of the file
	GetMethodQuery() *sitter.Query // Query to find the method nodes
	GetParamQuery() *sitter.Query  // Query to find the params of the method (depends on GetMethodQuery)
	GetCallQuery() *sitter.Query   // Query to find the calls that happen inside the method (depends on GetMethodQuery)

	// Logic hooks for language-specific data mapping
	MapMethod(tag, content string, m *model.Method)   // Maps the Method definition
	MapField(tag, content string, f *model.Field)     // Maps the field
	MapParam(tag, content string, p *model.Param)     // Maps the param for a method
	MapCall(tag, content string, c *model.MethodCall) // Maps the call inside a method

	//Language helper specific logic
	IsProjectImport(content string) bool
}

func GetAnalyser(l Language) (LanguageAnalyser, error) {
	switch l {
	case JAVA:
		return newJavaAnalyzer(), nil
	case TS:
		return newTypescriptAnalyser(), nil
	}

	return nil, &errors.LanguageNotFoundError{
		Arg:       fmt.Sprintf("%s", l),
		Message:   "Language not supported",
		Supported: []string{"JAVA", "TS"},
	}
}

type JavaAnalyzer struct {
	lang             *sitter.Language
	structQuery      *sitter.Query
	fieldQuery       *sitter.Query
	methodQuery      *sitter.Query
	paramQuery       *sitter.Query
	callQuery        *sitter.Query
	externalPrefixes []string //number of packages it need to match to count as project import
}

func newJavaAnalyzer() *JavaAnalyzer {
	l := java.GetLanguage()
	sQ, sErr := sitter.NewQuery([]byte(`(class_declaration name: (identifier) @class.name)
	(import_declaration (scoped_identifier) @import)`), l)
	fQ, fErr := sitter.NewQuery([]byte(`(field_declaration
				type: (_) @type
				declarator: (variable_declarator name: (identifier) @name))`), l)
	mQ, mErr := sitter.NewQuery([]byte(`(method_declaration
				type: (_) @method.return
				name: (identifier) @method.name) @m`), l)
	pQ, pErr := sitter.NewQuery([]byte(`(formal_parameter type: (_) @t name: (identifier) @n)`), l)
	cQ, cErr := sitter.NewQuery([]byte(`(method_invocation object: (_)? @tgt name: (identifier) @meth)`), l)

	if sErr != nil {
		panic(fmt.Errorf("Structure query could not be formed : [%s]", sErr.Error()))
	}

	if fErr != nil {
		panic(fmt.Errorf("Field query could not be formed : [%s]", fErr.Error()))
	}

	if mErr != nil {
		panic(fmt.Errorf("Method query could not be formed : [%s]", mErr.Error()))
	}

	if pErr != nil {
		panic(fmt.Errorf("Sub Method query Param Query could not be formed : [%s]", pErr.Error()))
	}

	if cErr != nil {
		panic(fmt.Errorf("Sub Method query Call Query could not be formed : [%s]", cErr.Error()))
	}

	return &JavaAnalyzer{
		lang:        l,
		structQuery: sQ,
		fieldQuery:  fQ,
		methodQuery: mQ,
		paramQuery:  pQ,
		callQuery:   cQ,
		externalPrefixes: []string{
			"java.",
			"javax.",
			"jakarta.",
			"org.springframework.",
			"com.google.",
			"org.apache.",
			"junit.",
		},
	}
}

func (j *JavaAnalyzer) GetLanguage() *sitter.Language { return java.GetLanguage() }

func (j *JavaAnalyzer) GetStructQuery() *sitter.Query {
	return j.structQuery
}

func (j *JavaAnalyzer) GetFieldQuery() *sitter.Query {
	return j.fieldQuery
}

func (j *JavaAnalyzer) GetMethodQuery() *sitter.Query {
	return j.methodQuery
}

func (j *JavaAnalyzer) GetParamQuery() *sitter.Query {
	return j.paramQuery
}

func (j *JavaAnalyzer) GetCallQuery() *sitter.Query {
	return j.callQuery
}

func (j *JavaAnalyzer) MapMethod(tag, content string, m *model.Method) {
	switch tag {
	case "method.return":
		m.ReturnType = content
	case "method.name":
		m.Name = content
	}
}

func (j *JavaAnalyzer) MapField(tag, content string, f *model.Field) {
	switch tag {
	case "type":
		f.Type = content
	case "name":
		f.Name = content
	}
}

func (j *JavaAnalyzer) MapParam(tag, content string, p *model.Param) {
	switch tag {
	case "t":
		p.Type = content
	case "n":
		p.Name = content
	}
}

func (j *JavaAnalyzer) MapCall(tag, content string, c *model.MethodCall) {
	switch tag {
	case "tgt":
		c.Target = content
	case "meth":
		c.Method = content
	}
}

func (j *JavaAnalyzer) IsProjectImport(content string) bool {
	//Probably a better way to do this than to hammer the prefixes
	for _, prefix := range j.externalPrefixes {
		if strings.HasPrefix(content, prefix) {
			return false
		}
	}
	return true
}

type TypeScriptAnalyzer struct {
	lang        *sitter.Language
	structQuery *sitter.Query
	fieldQuery  *sitter.Query
	methodQuery *sitter.Query
	paramQuery  *sitter.Query
	callQuery   *sitter.Query
}

func newTypescriptAnalyser() *TypeScriptAnalyzer {
	l := typescript.GetLanguage()
	sQ, _ := sitter.NewQuery([]byte(`(class_declaration name: (type_identifier) @class.name)
	(import_statement (import_clause (named_imports (import_specifier name: (identifier) @import))))`), l)
	fQ, _ := sitter.NewQuery([]byte(`(public_field_definition
		name: (property_identifier) @name
		type: (type_annotation (type_identifier) @type))`), l)
	// In TS, return type is inside a type_annotation after the parameters
	mQ, _ := sitter.NewQuery([]byte(`(method_definition name: (property_identifier) @method.name return_type: (type_annotation)? @method.return) @m`), l)
	pQ, _ := sitter.NewQuery([]byte(`(formal_parameter name: (identifier) @n type: (type_annotation (type_identifier) @t)?)`), l)
	// TS calls look like object.method() or just method()
	cQ, _ := sitter.NewQuery([]byte(`(call_expression function: (member_expression object: (identifier) @tgt property: (property_identifier) @meth))`), l)

	return &TypeScriptAnalyzer{
		lang:        l,
		structQuery: sQ,
		fieldQuery:  fQ,
		methodQuery: mQ,
		paramQuery:  pQ,
		callQuery:   cQ,
	}
}

func (t *TypeScriptAnalyzer) GetLanguage() *sitter.Language { return typescript.GetLanguage() }

func (t *TypeScriptAnalyzer) GetStructQuery() *sitter.Query {
	return t.structQuery
}

func (t *TypeScriptAnalyzer) GetFieldQuery() *sitter.Query {
	return t.fieldQuery
}

func (t *TypeScriptAnalyzer) GetMethodQuery() *sitter.Query {
	// In TS, return type is inside a type_annotation after the parameters
	return t.methodQuery
}

func (t *TypeScriptAnalyzer) GetParamQuery() *sitter.Query {
	return t.paramQuery
}

func (t *TypeScriptAnalyzer) GetCallQuery() *sitter.Query {
	// TS calls look like object.method() or just method()
	return t.callQuery
}

func (t *TypeScriptAnalyzer) MapMethod(tag, content string, m *model.Method) {
	switch tag {
	case "method.name":
		m.Name = content
	case "method.return":
		m.ReturnType = strings.TrimSpace(strings.TrimPrefix(content, ":")) // Remove the ":" from ": string"
	}
}

func (t *TypeScriptAnalyzer) MapField(tag, content string, f *model.Field) {
	switch tag {
	case "name":
		f.Name = content
	case "type":
		f.Type = content
	}
}

func (t *TypeScriptAnalyzer) MapParam(tag, content string, p *model.Param) {
	switch tag {
	case "n":
		p.Name = content
	case "t":
		p.Type = content
	}
}

func (t *TypeScriptAnalyzer) MapCall(tag, content string, c *model.MethodCall) {
	switch tag {
	case "tgt":
		c.Target = content
	case "meth":
		c.Method = content
	}
}

func (t *TypeScriptAnalyzer) IsProjectImport(content string) bool {
	// In TS, local files are almost always relative paths
	return strings.HasPrefix(content, ".") || strings.HasPrefix(content, "..")
}
