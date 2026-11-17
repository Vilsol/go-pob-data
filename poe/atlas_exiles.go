package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type AtlasExile struct {
	raw.AtlasExile
}

var AtlasExiles []*AtlasExile

func InitializeAtlasExiles(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "AtlasExiles", &AtlasExiles, nil, assetCache)
}
