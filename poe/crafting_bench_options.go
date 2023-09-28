package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type CraftingBenchOption struct {
	raw.CraftingBenchOption
}

var CraftingBenchOptions []*CraftingBenchOption

func InitializeCraftingBenchOptions(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CraftingBenchOptions", &CraftingBenchOptions, nil, assetCache)
}
