package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Stat struct {
	raw.Stat
}

var Stats []*Stat

func InitializeStats(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Stats", &Stats, nil, assetCache)
}
