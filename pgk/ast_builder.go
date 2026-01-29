package pgk

type Property struct {
	Name string
	Type string
}

type Method struct {
	Name       string
	Params     []Property
	ReturnType string
	Exceptions []string
}

type FileAst struct {
	FileName             string
	FileDir              string
	DependencyReferences map[string]int
	Properties           []Property
	Methods              []Method
}

func BuildAstTree(file *FileScan) (*FileAst, error) {

	ast := &FileAst{
		FileName: file.name,
		FileDir:  file.dir,
	}

	for line := range file.lines {

	}

	return ast, nil
}
