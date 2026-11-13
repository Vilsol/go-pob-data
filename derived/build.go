package derived

import (
	"context"
	"fmt"

	"github.com/tinylib/msgp/msgp"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe2/pobdata"
)

// Asset is one table's rows as its published asset holds them before compression.
type Asset struct {
	// Path is the asset's AssetPath.
	Path    string
	Encoded []byte
	Rows    int
}

type table struct {
	build func(source) (Asset, error)
	name  string
	// tables and statFiles are the raw tables and stat description files the table is derived from.
	tables    []string
	statFiles []string
}

func encode[T Row, PT interface {
	*T
	msgp.Encodable
}](build func(source) ([]T, error)) func(source) (Asset, error) {
	return func(src source) (Asset, error) {
		rows, err := build(src)
		if err != nil {
			return Asset{}, err
		}
		encoded, err := pobdata.Encode[T, PT](rows)
		if err != nil {
			return Asset{}, err
		}
		return Asset{Path: AssetPath[T](encoded), Encoded: encoded, Rows: len(rows)}, nil
	}
}

var tables = []table{
	{
		name:      ModScalabilityTable,
		tables:    []string{"Stats"},
		statFiles: []string{"stat_descriptions"},
		build:     encode(buildModScalability),
	},
	{
		name:      ItemModStatOrderTable,
		tables:    []string{"Stats", "ModFamily", "BuffTemplates", "Mods"},
		statFiles: []string{"stat_descriptions", "tincture_stat_descriptions", "graft_stat_descriptions"},
		build:     encode(buildItemModStatOrder),
	},
}

// Build derives every table whose inputs version publishes, read through http.DefaultClient and
// assetCache. It sets the poe package's tables, as their Initialize functions do.
func Build(ctx context.Context, version string, assetCache loader.AssetCache) ([]Asset, error) {
	manifest, err := loader.LoadManifest(ctx, version)
	if err != nil {
		return nil, err
	}
	if manifest.JSONOnly {
		return nil, nil
	}

	src := source{ctx: ctx, version: version, assetCache: assetCache}
	var assets []Asset
	for _, t := range tables {
		if !hasInputs(manifest, t) {
			continue
		}
		asset, err := t.build(src)
		if err != nil {
			return nil, fmt.Errorf("%s of %s: %w", t.name, version, err)
		}
		assets = append(assets, asset)
	}
	return assets, nil
}

func hasInputs(manifest *loader.Manifest, t table) bool {
	for _, name := range t.tables {
		if !manifest.HasTable(name) {
			return false
		}
	}
	for _, name := range t.statFiles {
		if !manifest.HasStatTranslation(name) {
			return false
		}
	}
	return true
}
