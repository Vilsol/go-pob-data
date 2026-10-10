package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ModFamily struct {
	raw.ModFamily
}

var ModFamilies []*ModFamily

var ModFamiliesByID map[string]*ModFamily

func InitializeModFamilies(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ModFamily", &ModFamilies, func(count int64) {
		ModFamiliesByID = make(map[string]*ModFamily, count)
	}, assetCache, func(obj *ModFamily) {
		ModFamiliesByID[obj.ID] = obj
	})
}
