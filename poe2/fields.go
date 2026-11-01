package poe2

import (
	"embed"
	"reflect"
	"slices"
	"strings"

	"github.com/Vilsol/go-pob-data/fields"
)

// fields/<version>.json are go-pob's field lists of these rows, as fields/ holds PoE1's.
//
//go:embed fields/*.json
var embeddedFields embed.FS

// RowType is the type the loader decodes asset ("raw/<table>") into, for a fields.List.
func RowType(asset string) (reflect.Type, bool) {
	name, ok := strings.CutPrefix(asset, "raw/")
	if !ok {
		return nil, false
	}
	t, ok := Tables[name]
	return t, ok
}

var fieldLists = func() map[string]*fields.List {
	files, err := embeddedFields.ReadDir("fields")
	if err != nil {
		panic(err)
	}

	out := map[string]*fields.List{}
	for _, file := range files {
		b, err := embeddedFields.ReadFile("fields/" + file.Name())
		if err != nil {
			panic(err)
		}

		// TestEmbeddedFieldLists keeps these valid.
		l, err := fields.ParseFor(b, RowType)
		if err != nil {
			panic(file.Name() + ": " + err.Error())
		}
		if l.Version+".json" != file.Name() {
			panic(file.Name() + ": list of version " + l.Version)
		}
		out[l.Version] = l
	}
	return out
}()

// FieldVersions are the game versions that have a field list, sorted.
func FieldVersions() []string {
	versions := make([]string, 0, len(fieldLists))
	for version := range fieldLists {
		versions = append(versions, version)
	}
	slices.Sort(versions)
	return versions
}

// FieldList returns the field list of version.
func FieldList(version string) (*fields.List, bool) {
	l, ok := fieldLists[version]
	return l, ok
}

// MinVariant is fields.MinVariant for these rows: the variant of asset's .min in version, if version's
// list has asset. It fits loader.VariantFunc.
func MinVariant(version string, asset string) (string, bool) {
	l, ok := fieldLists[version]
	if !ok {
		return "", false
	}
	return l.MinVariant(asset)
}
