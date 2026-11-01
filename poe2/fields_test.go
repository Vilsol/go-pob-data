package poe2

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
)

func TestEmbeddedFieldLists(t *testing.T) {
	if len(FieldVersions()) == 0 {
		t.Fatal("no embedded lists")
	}
	for _, version := range FieldVersions() {
		l, _ := FieldList(version)
		for asset, a := range l.Assets {
			variant, ok := MinVariant(version, asset)
			if !ok || !strings.HasPrefix(variant, "min-") || len(variant) != len("min-")+8 {
				t.Errorf("%s %s: variant %q", version, asset, variant)
			}
			if rt, _ := RowType(asset); rt.Name() != a.Type {
				t.Errorf("%s %s: type %s, Tables has %s", version, asset, a.Type, rt.Name())
			}
		}
	}

	if _, ok := MinVariant(version, "raw/NoSuchTable"); ok {
		t.Error("a variant for an asset the list does not have")
	}
	if _, ok := fields.MinVariant(version, "raw/Mods"); ok {
		t.Error("the PoE1 lists have a PoE2 version")
	}
}

// Every listed asset has a committed .min that is the projection of its committed full asset under the list.
func TestMinMatchesList(t *testing.T) {
	for _, version := range FieldVersions() {
		l, _ := FieldList(version)
		for _, asset := range slices.Sorted(func(yield func(string) bool) {
			for asset := range l.Assets {
				if !yield(asset) {
					return
				}
			}
		}) {
			t.Run(version+"/"+asset, func(t *testing.T) {
				rt, _ := RowType(asset)
				dir := filepath.Join("..", "data", version)
				full := mustRead(t, filepath.Join(dir, loader.AssetPath(asset, loader.TupleVariant(reflect.New(rt).Interface()))))
				want, err := fields.ProjectType(rt, asset, l.Assets[asset].Keep, full)
				if err != nil {
					t.Fatal(err)
				}
				variant, _ := l.MinVariant(asset)
				got := mustRead(t, filepath.Join(dir, loader.AssetPath(asset, variant)))
				if !bytes.Equal(got, want) {
					t.Fatalf("the .min is stale: run mise run project-min")
				}
				if fullRows, minRows := rowCount(t, full), rowCount(t, got); fullRows != minRows || minRows == 0 {
					t.Fatalf("the .min has %d rows, the full asset %d", minRows, fullRows)
				}
			})
		}
	}
}

// Every committed .min of a PoE2 version is of a listed asset.
func TestMinAssetsListed(t *testing.T) {
	mins, err := filepath.Glob(filepath.Join("..", "data", "0.*", "raw", "*.min-*.msgpack.br"))
	if err != nil {
		t.Fatal(err)
	}
	if len(mins) == 0 {
		t.Fatal("no .min assets")
	}
	for _, path := range mins {
		version := filepath.Base(filepath.Dir(filepath.Dir(path)))
		name, variant, _ := strings.Cut(strings.TrimSuffix(filepath.Base(path), ".msgpack.br"), ".")
		l, ok := FieldList(version)
		if !ok {
			t.Errorf("%s: %s has no list", path, version)
			continue
		}
		if current, ok := l.MinVariant("raw/" + name); !ok {
			t.Errorf("%s is not of a listed asset", path)
		} else if variant != current {
			t.Logf("%s is of an older list revision", path)
		}
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func rowCount(t *testing.T, asset []byte) int64 {
	t.Helper()
	n, err := msgp.NewReader(brotli.NewReader(bytes.NewReader(asset))).ReadInt64()
	if err != nil {
		t.Fatal(err)
	}
	return n
}
