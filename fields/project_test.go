package fields

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
)

const dataDir = "../data"

// readAsset returns nil for an asset that does not exist.
func readAsset(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	if bytes.HasPrefix(b, []byte("version https://git-lfs")) {
		t.Skip("git lfs pointer, run git lfs pull")
	}

	return b
}

func fullAsset(t *testing.T, version string, asset string) []byte {
	t.Helper()

	rowType, _ := RowType(asset)
	return readAsset(t, filepath.Join(dataDir, version, loader.AssetPath(asset, loader.TupleVariant(reflect.New(rowType).Interface()))))
}

// shipped tells a table the version does not ship, which the loader reads as empty, from a missing asset.
func shipped(version string, asset string) bool {
	_, err := os.Stat(filepath.Join(dataDir, version, asset+".json.gz"))
	return !os.IsNotExist(err) || !strings.HasPrefix(asset, "raw/")
}

// decode decodes an asset the way the loader does: pointers to rows (one TranslationFile for a
// translation), then EOF.
func decode(t *testing.T, asset string, b []byte) []reflect.Value {
	t.Helper()

	rowType, _ := RowType(asset)
	r := msgp.NewReader(brotli.NewReader(bytes.NewReader(b)))
	count := int64(1)
	if strings.HasPrefix(asset, "raw/") {
		var err error
		if count, err = r.ReadInt64(); err != nil {
			t.Fatal(err)
		}
	}

	rows := make([]reflect.Value, count)
	for i := range rows {
		rows[i] = reflect.New(rowType)
		if err := rows[i].Interface().(msgp.Decodable).DecodeMsg(r); err != nil {
			t.Fatalf("row %d: %v", i, err)
		}
	}
	if _, err := r.R.ReadByte(); !errors.Is(err, io.EOF) {
		t.Fatal("trailing bytes after the last row")
	}

	return rows
}

func isZero(v reflect.Value) bool {
	return v.IsZero() || ((v.Kind() == reflect.Map || v.Kind() == reflect.Slice) && v.Len() == 0)
}

// projected checks minAsset against full independently of Project: a listed path with no listed children
// holds full's value, an unlisted field is zero, and pointers and slices on a listed path match full's
// nil-ness and length.
func projected(at string, full reflect.Value, minAsset reflect.Value, n *node) error {
	switch full.Kind() {
	case reflect.Pointer:
		if full.IsNil() != minAsset.IsNil() {
			return errors.Errorf("%s: nil in only one of full and .min", at)
		}
		if full.IsNil() {
			return nil
		}
		return projected(at, full.Elem(), minAsset.Elem(), n)
	case reflect.Slice, reflect.Array:
		if full.Len() != minAsset.Len() {
			return errors.Errorf("%s: %d elements in full, %d in .min", at, full.Len(), minAsset.Len())
		}
		for i := range full.Len() {
			if err := projected(fmt.Sprintf("%s[%d]", at, i), full.Index(i), minAsset.Index(i), n); err != nil {
				return err
			}
		}
		return nil
	case reflect.Struct:
		for i := range full.NumField() {
			name := full.Type().Field(i).Name
			path := at + "." + name
			c := n.children[name]
			switch {
			case c == nil:
				if !isZero(minAsset.Field(i)) {
					return errors.Errorf("%s is not listed, but not zero in .min", path)
				}
			case len(c.children) == 0:
				if !reflect.DeepEqual(full.Field(i).Interface(), minAsset.Field(i).Interface()) {
					return errors.Errorf("%s is listed, but differs in .min", path)
				}
			default:
				if err := projected(path, full.Field(i), minAsset.Field(i), c); err != nil {
					return err
				}
			}
		}
		return nil
	default:
		return errors.Errorf("%s: cannot descend into %s", at, full.Type())
	}
}

func assertRoundTrip(t *testing.T, asset string, keep []string, full []byte, minAsset []byte) {
	t.Helper()

	rowType, _ := RowType(asset)
	tree, err := keepTree(rowType, keep)
	if err != nil {
		t.Fatal(err)
	}

	fullRows, minRows := decode(t, asset, full), decode(t, asset, minAsset)
	if len(fullRows) != len(minRows) {
		t.Fatalf("%d rows in full, %d in .min", len(fullRows), len(minRows))
	}
	for i := range fullRows {
		if err := projected(fmt.Sprintf("row %d", i), fullRows[i], minRows[i], tree); err != nil {
			t.Fatal(err)
		}
	}
}

// Decoding a committed .min yields the full asset's value at every listed path, zero everywhere else,
// and the same row count.
func TestMinRoundTrip(t *testing.T) {
	for _, version := range Versions() {
		l, _ := For(version)
		for asset, a := range l.Assets {
			t.Run(version+"/"+asset, func(t *testing.T) {
				if !shipped(version, asset) {
					t.Skip("not in this version")
				}
				variant, _ := l.MinVariant(asset)
				minAsset := readAsset(t, filepath.Join(dataDir, version, loader.AssetPath(asset, variant)))
				if minAsset == nil {
					t.Fatal("no .min asset")
				}
				assertRoundTrip(t, asset, a.Keep, fullAsset(t, version, asset), minAsset)
			})
		}
	}
}

// Projecting the committed full assets with the committed lists reproduces every committed .min, and
// every .min on disk is of a list revision (the current one, or an older one kept for older consumers).
func TestMinMatchesList(t *testing.T) {
	for _, version := range Versions() {
		l, _ := For(version)
		for asset, a := range l.Assets {
			t.Run(version+"/"+asset, func(t *testing.T) {
				variant, _ := l.MinVariant(asset)
				minPath := filepath.Join(dataDir, version, loader.AssetPath(asset, variant))
				committed := readAsset(t, minPath)

				full := fullAsset(t, version, asset)
				if full == nil {
					if shipped(version, asset) {
						t.Fatalf("%s is in the list, but has no full asset", asset)
					}
					if committed != nil {
						t.Errorf("%s exists for a table this version does not ship", minPath)
					}
					t.Skip("not in this version")
				}

				want, err := Project(asset, a.Keep, full)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(want, committed) {
					t.Errorf("%s is missing or stale: run mise run project-min", minPath)
				}
			})
		}
	}

	rawMins, err := filepath.Glob(filepath.Join(dataDir, "*", "raw", "*.min-*.msgpack.br"))
	if err != nil {
		t.Fatal(err)
	}
	translationMins, err := filepath.Glob(filepath.Join(dataDir, "*", "stat_translations", "*", "*.min-*.msgpack.br"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range append(rawMins, translationMins...) {
		rel, err := filepath.Rel(dataDir, path)
		if err != nil {
			t.Fatal(err)
		}
		version, file, _ := strings.Cut(filepath.ToSlash(rel), "/")
		asset, variant, _ := strings.Cut(strings.TrimSuffix(file, ".msgpack.br"), ".")

		l, ok := For(version)
		if !ok {
			t.Errorf("%s: %s has no list", path, version)
			continue
		}
		if _, known := RowType(asset); !known {
			t.Errorf("%s is not of a known asset", path)
			continue
		}
		if current, _ := l.MinVariant(asset); variant != current {
			t.Logf("%s is of an older list revision", path)
		}
	}
}

// A list revision needs only the projector and the committed full assets: adding and removing a path
// in memory gives a new variant, and a .min with the added field restored and the removed one zeroed.
func TestListRevisionReprojects(t *testing.T) {
	const version, asset = "3.29", "raw/WeaponTypes"
	l, _ := For(version)
	full := fullAsset(t, version, asset)
	if full == nil {
		t.Fatal("no full asset")
	}
	rows := decode(t, asset, full)

	// Pick fields that hold data in some row, so neither edit is invisible.
	holdsData := func(field string) bool {
		for _, row := range rows {
			if !row.Elem().FieldByName(field).IsZero() {
				return true
			}
		}
		return false
	}
	var added, removed string
	rowType, _ := RowType(asset)
	for i := range rowType.NumField() {
		name := rowType.Field(i).Name
		switch {
		case !holdsData(name):
		case slices.Contains(l.Assets[asset].Keep, name) && removed == "":
			removed = name
		case !slices.Contains(l.Assets[asset].Keep, name) && added == "":
			added = name
		}
	}
	if added == "" || removed == "" {
		t.Fatalf("no field to add (%q) or remove (%q)", added, removed)
	}

	keep := append(slices.DeleteFunc(slices.Clone(l.Assets[asset].Keep), func(p string) bool { return p == removed }), added)
	revised := &List{Format: Format, Version: version, Assets: map[string]Asset{asset: {Type: l.Assets[asset].Type, Keep: keep}}}

	before, _ := l.MinVariant(asset)
	after, _ := revised.MinVariant(asset)
	if before == after {
		t.Fatalf("the revised list keeps variant %s", before)
	}

	minAsset, err := Project(asset, keep, full)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(minAsset, readAsset(t, filepath.Join(dataDir, version, loader.AssetPath(asset, before)))) {
		t.Fatal("the revised .min is the committed one")
	}
	assertRoundTrip(t, asset, keep, full, minAsset)

	for _, row := range decode(t, asset, minAsset) {
		if !row.Elem().FieldByName(removed).IsZero() {
			t.Fatalf("%s is removed from the list, but kept", removed)
		}
	}
	t.Logf("added %s, removed %s: %s -> %s", added, removed, before, after)
}

func TestProjectRejectsPlantedPath(t *testing.T) {
	full := fullAsset(t, "3.29", "raw/WeaponTypes")
	_, err := Project("raw/WeaponTypes", []string{"Critical", "NoSuchField"}, full)
	want := `raw/WeaponTypes: key path "NoSuchField": WeaponType has no field NoSuchField`
	if err == nil || err.Error() != want {
		t.Errorf("got %v\nwant %s", err, want)
	}
}

// "A.B" keeps B inside A through slices and pointers, and a path with listed children keeps only them.
func TestProjectNestedPaths(t *testing.T) {
	const asset = "stat_translations/en/stat_descriptions"
	full := fullAsset(t, "3.29", asset)
	keep := []string{"Descriptors", "Descriptors.List", "Descriptors.List.Conditions.Min", "Descriptors.List.String"}

	minAsset, err := Project(asset, keep, full)
	if err != nil {
		t.Fatal(err)
	}
	assertRoundTrip(t, asset, keep, full, minAsset)
}
