package languages

import (
	"fmt"
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
	GetStructQuery() string // Query to find the struct of the file (class name & imports)
	GetFieldQuery() string  // Query to find the fields of the file
	GetMethodQuery() string // Query to find the method nodes
	GetParamQuery() string  // Query to find the params of the method (depends on GetMethodQuery)
	GetCallQuery() string   // Query to find the calls that happen inside the method (depends on GetMethodQuery)

	// Logic hooks for language-specific data mapping
	MapMethod(tag, content string) Method   // Maps the Method definition
	MapField(tag, content string) Field     // Maps the field
	MapParam(tag, content string) Param     // Maps the param for a method
	MapCall(tag, content string) MethodCall // Maps the call inside a method
}

func GetAnalyser(l Language) (LanguageAnalyser, error) {
	switch l {
	case JAVA:
		return &JavaAnalyzer{}, nil
	case TS:
		return &TypeScriptAnalyzer{}, nil
	}

	return nil, &errors.LanguageNotFoundError{
		Arg:       fmt.Sprintf("%s", l),
		Message:   "Language not supported",
		Supported: []string{"JAVA", "TS"},
	}
}

type JavaAnalyzer struct{}

func (j JavaAnalyzer) GetLanguage() *sitter.Language { return java.GetLanguage() }

func (j JavaAnalyzer) GetStructQuery() string {
	return `
		(class_declaration name: (identifier) @class.name)
		(import_declaration (scoped_identifier) @import)
	`
}

func (j JavaAnalyzer) GetFieldQuery() string {
	return `
		(field_declaration
			type: [(type_identifier) (primitive_type)] @type
			declarator: (variable_declarator name: (identifier) @name))
	`
}

func (j JavaAnalyzer) GetMethodQuery() string {
	return `(method_declaration type: [(type_identifier) (primitive_type)] @method.return name: (identifier) @method.name) @m`
}

func (j JavaAnalyzer) GetParamQuery() string {
	return `(formal_parameter type: [(type_identifier) (primitive_type)] @t name: (identifier) @n)`
}

func (j JavaAnalyzer) GetCallQuery() string {
	return `(method_invocation object: (_)? @tgt name: (identifier) @meth)`
}

func (j JavaAnalyzer) MapMethod(tag, content string) Method {
	m := Method{}
	if tag == "method.return" {
		m.ReturnType = content
	}
	if tag == "method.name" {
		m.Name = content
	}

	return m
}

func (j JavaAnalyzer) MapField(tag, content string) Field {
	f := Field{}
	if tag == "type" {
		f.Type = content
	}
	if tag == "name" {
		f.Name = content
	}

	return f
}

func (j JavaAnalyzer) MapParam(tag, content string) Param {
	p := Param{}
	if tag == "t" {
		p.Type = content
	}
	if tag == "n" {
		p.Name = content
	}

	return p
}

func (j JavaAnalyzer) MapCall(tag, content string) MethodCall {
	c := MethodCall{}
	if tag == "tgt" {
		c.Target = content
	}
	if tag == "meth" {
		c.Method = content
	}

	return c
}

type TypeScriptAnalyzer struct{}

func (t TypeScriptAnalyzer) GetLanguage() *sitter.Language { return typescript.GetLanguage() }

func (t TypeScriptAnalyzer) GetStructQuery() string {
	return `
		(class_declaration name: (type_identifier) @class.name)
		(import_statement (import_clause (named_imports (import_specifier name: (identifier) @import))))
	`
}

func (t TypeScriptAnalyzer) GetFieldQuery() string {
	return `
		(public_field_definition
			name: (property_identifier) @name
			type: (type_annotation (type_identifier) @type))
	`
}

func (t TypeScriptAnalyzer) GetMethodQuery() string {
	// In TS, return type is inside a type_annotation after the parameters
	return `(method_definition name: (property_identifier) @method.name return_type: (type_annotation)? @method.return) @m`
}

func (t TypeScriptAnalyzer) GetParamQuery() string {
	return `(formal_parameter name: (identifier) @n type: (type_annotation (type_identifier) @t)?)`
}

func (t TypeScriptAnalyzer) GetCallQuery() string {
	// TS calls look like object.method() or just method()
	return `(call_expression function: (member_expression object: (identifier) @tgt property: (property_identifier) @meth))`
}

func (t TypeScriptAnalyzer) MapMethod(tag, content string) Method {
	m := Method{}
	if tag == "method.name" {
		m.Name = content
	}
	if tag == "method.return" {
		// Remove the ":" from ": string"
		m.ReturnType = strings.TrimSpace(strings.TrimPrefix(content, ":"))
	}

	return m
}

func (t TypeScriptAnalyzer) MapField(tag, content string) Field {
	f := Field{}
	if tag == "name" {
		f.Name = content
	}
	if tag == "type" {
		f.Type = content
	}

	return f
}

func (t TypeScriptAnalyzer) MapParam(tag, content string) Param {
	p := Param{}
	if tag == "n" {
		p.Name = content
	}
	if tag == "t" {
		p.Type = content
	}

	return p
}

func (t TypeScriptAnalyzer) MapCall(tag, content string) MethodCall {
	c := MethodCall{}
	if tag == "tgt" {
		c.Target = content
	}
	if tag == "meth" {
		c.Method = content
	}

	return c
}
