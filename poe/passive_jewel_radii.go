package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveJewelRadius struct {
	raw.PassiveJewelRadius
}

var PassiveJewelRadii []*PassiveJewelRadius

func InitializePassiveJewelRadii(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveJewelRadii", &PassiveJewelRadii, nil, assetCache)
}
