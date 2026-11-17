package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type FishingLureType struct {
	raw.FishingLureType
}

var FishingLureTypes []*FishingLureType

func InitializeFishingLureTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "FishingLureTypes", &FishingLureTypes, nil, assetCache)
}
