package poe2

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

// A struct left out of //msgp:tuple is map-encoded inside a tuple asset, and a row type left out of
// Tables has no asset.
func TestTupleCoversAllStructs(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	var structs, tuple, ignored []string
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
				if rest, ok := strings.CutPrefix(c.Text, "//msgp:ignore "); ok {
					ignored = append(ignored, strings.Fields(rest)...)
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

	listed := slices.Sorted(slices.Values(slices.Concat(tuple, ignored)))
	slices.Sort(structs)
	if !slices.Equal(structs, listed) {
		t.Errorf("//msgp:tuple and //msgp:ignore must list exactly the structs of poe2\nstructs: %v\nlisted:  %v", structs, listed)
	}

	var rows []string
	for _, rt := range Tables {
		rows = append(rows, rt.Name())
	}
	slices.Sort(rows)
	slices.Sort(tuple)
	if !slices.Equal(rows, tuple) {
		t.Errorf("Tables must name every tuple row once\nTables: %v\ntuple:  %v", rows, tuple)
	}
}

// Every table's tuple asset decodes to the rows its json decodes to.
func TestTupleMatchesJSON(t *testing.T) {
	for _, name := range slices.Sorted(func(yield func(string) bool) {
		for name := range Tables {
			if !yield(name) {
				return
			}
		}
	}) {
		rt := Tables[name]
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join("..", "data", version)
			fromJSON := reflect.New(reflect.SliceOf(rt))
			if err := json.Unmarshal(readCompressed(t, filepath.Join(dir, "raw", name+".json.gz")), fromJSON.Interface()); err != nil {
				t.Fatal(err)
			}

			asset := filepath.Join(dir, loader.AssetPath("raw/"+name, loader.TupleVariant(reflect.New(rt).Interface())))
			r := msgp.NewReader(bytes.NewReader(readCompressed(t, asset)))
			count, err := r.ReadInt64()
			if err != nil {
				t.Fatal(err)
			}
			fromTuple := reflect.MakeSlice(reflect.SliceOf(rt), int(count), int(count))
			for i := range fromTuple.Len() {
				if err := fromTuple.Index(i).Addr().Interface().(msgp.Decodable).DecodeMsg(r); err != nil {
					t.Fatalf("row %d: %v", i, err)
				}
			}
			if _, err := r.R.ReadByte(); err != io.EOF {
				t.Fatal("trailing bytes after the last row")
			}

			want := fromJSON.Elem()
			if want.Len() == 0 {
				t.Fatal("no rows")
			}
			// An empty list decodes to nil from a tuple and to an empty slice from json.
			nilEmptySlices(want)
			if !reflect.DeepEqual(want.Interface(), fromTuple.Interface()) {
				for i := range want.Len() {
					if !reflect.DeepEqual(want.Index(i).Interface(), fromTuple.Index(i).Interface()) {
						t.Fatalf("row %d: json %+v, tuple %+v", i, want.Index(i).Interface(), fromTuple.Index(i).Interface())
					}
				}
				t.Fatalf("json has %d rows, tuple %d", want.Len(), fromTuple.Len())
			}
		})
	}
}

// LoadRaw reads the published tuple asset of a row type by its layout.
func TestLoadRaw(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	rows, err := loader.LoadRaw[*Mod](context.Background(), version, ModsTable, nil, nopCache{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	for i, row := range rows {
		if row.Key != i {
			t.Fatalf("row %d has key %d", i, row.Key)
		}
	}
	if misses := cdn.Misses(); len(misses) > 0 {
		t.Fatalf("misses: %v", misses)
	}
}

func readCompressed(t *testing.T, path string) []byte {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	var r io.Reader = brotli.NewReader(f)
	if strings.HasSuffix(path, ".gz") {
		if r, err = gzip.NewReader(f); err != nil {
			t.Fatal(err)
		}
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func nilEmptySlices(v reflect.Value) {
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			nilEmptySlices(v.Elem())
		}
	case reflect.Slice:
		if v.Len() == 0 && v.CanSet() {
			v.SetZero()
			return
		}
		for i := range v.Len() {
			nilEmptySlices(v.Index(i))
		}
	case reflect.Struct:
		for i := range v.NumField() {
			nilEmptySlices(v.Field(i))
		}
	default:
	}
}
