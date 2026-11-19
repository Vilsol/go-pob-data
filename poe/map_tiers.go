package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MapTier struct {
	raw.MapTier
}

var MapTiers []*MapTier

var MapTiersByTier map[int]*MapTier

func InitializeMapTiers(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MapTiers", &MapTiers, func(count int64) {
		MapTiersByTier = make(map[int]*MapTier, count)
	}, assetCache, func(obj *MapTier) {
		MapTiersByTier[obj.Tier] = obj
	})
}
