package raw

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// A struct left out of //msgp:tuple is map-encoded, inside a tuple asset whose name does not say so.
func TestTupleCoversAllStructs(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	var structs, tuple []string
	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_gen.go") {
			continue
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}

		for _, group := range f.Comments {
			for _, c := range group.List {
				if strings.HasPrefix(c.Text, "//msgp:vartuple") {
					t.Errorf("%s: vartuple misassigns fields when one is inserted, use tuple", path)
				}
				if rest, ok := strings.CutPrefix(c.Text, "//msgp:tuple "); ok {
					tuple = append(tuple, strings.Fields(rest)...)
				}
			}
		}

		ast.Inspect(f, func(n ast.Node) bool {
			if spec, ok := n.(*ast.TypeSpec); ok {
				if _, ok := spec.Type.(*ast.StructType); ok {
					structs = append(structs, spec.Name.Name)
				}
			}
			return true
		})
	}

	slices.Sort(structs)
	slices.Sort(tuple)
	if !slices.Equal(structs, tuple) {
		t.Errorf("//msgp:tuple must list exactly the structs of raw\nstructs: %v\ntuple:   %v", structs, tuple)
	}
}
