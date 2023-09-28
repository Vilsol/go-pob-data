package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type CostType struct {
	raw.CostType
}

var CostTypes []*CostType

func InitializeCostTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CostTypes", &CostTypes, nil, assetCache)
}
