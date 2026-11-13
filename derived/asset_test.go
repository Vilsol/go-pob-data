package derived

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

// Every committed PoE1 version's derived assets are what Build derives from its committed data now: one
// per table whose inputs it publishes, and no other.
func TestAssetsCurrent(t *testing.T) {
	testdata.ServeCommittedData(t)
	dirs, err := filepath.Glob(filepath.Join("..", "data", "3.*"))
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		version := filepath.Base(dir)
		t.Run(version, func(t *testing.T) {
			assets, err := Build(context.Background(), version, noCache{})
			if err != nil {
				t.Fatal(err)
			}
			want := map[string][]byte{}
			for _, asset := range assets {
				want[asset.Path] = asset.Encoded
			}

			published, err := filepath.Glob(filepath.Join(dir, "derived", "*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(published) != len(want) {
				t.Errorf("%d assets published, %d derived: run mise run derive", len(published), len(want))
			}
			for _, path := range published {
				rel := "derived/" + filepath.Base(path)
				compressed, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if bytes.HasPrefix(compressed, []byte("version https://git-lfs")) {
					t.Fatalf("%s is a git lfs pointer, run git lfs pull", path)
				}
				encoded, err := io.ReadAll(brotli.NewReader(bytes.NewReader(compressed)))
				if err != nil {
					t.Fatalf("%s: %v", path, err)
				}
				if w, ok := want[rel]; !ok || !bytes.Equal(encoded, w) {
					t.Errorf("%s is not what its version derives: run mise run derive", rel)
				}
			}
		})
	}
}

// A struct left out of //msgp:tuple is map-encoded inside a tuple asset.
func TestEveryRowIsTuple(t *testing.T) {
	gen, err := os.ReadFile("derived_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(gen, []byte("ReadMapHeader")) {
		t.Error("derived_gen.go decodes a map-encoded struct: list it in //msgp:tuple")
	}
}

// Load reads the asset a manifest names for the row type's layout, and a version without one is
// ErrAssetNotFound.
func TestLoad(t *testing.T) {
	testdata.ServeCommittedData(t)
	rows, err := Load[ModScalability](context.Background(), pobVersion, noCache{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) < 10000 || rows[0].Line != "" || !strings.Contains(rows[1].Line, "#") {
		t.Errorf("%d rows, the first %+v", len(rows), rows[:min(2, len(rows))])
	}

	_, err = Load[ModScalability](context.Background(), "0.5", noCache{})
	if !errors.Is(err, loader.ErrAssetNotFound) {
		t.Errorf("0.5 loads with %v, want ErrAssetNotFound", err)
	}
}
