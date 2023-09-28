package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Flask struct {
	raw.Flask
}

var Flasks []*Flask

func InitializeFlasks(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Flasks", &Flasks, nil, assetCache)
}
