package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ComponentCharge struct {
	raw.ComponentCharge
}

var ComponentCharges []*ComponentCharge

func InitializeComponentCharges(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ComponentCharges", &ComponentCharges, nil, assetCache)
}
