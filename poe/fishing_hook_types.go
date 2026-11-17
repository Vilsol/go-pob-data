package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type FishingHookType struct {
	raw.FishingHookType
}

var FishingHookTypes []*FishingHookType

func InitializeFishingHookTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "FishingHookTypes", &FishingHookTypes, nil, assetCache)
}
