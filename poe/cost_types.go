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

var CostTypesByID map[string]*CostType

func InitializeCostTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CostTypes", &CostTypes, func(count int64) {
		CostTypesByID = make(map[string]*CostType, count)
	}, assetCache, func(obj *CostType) {
		CostTypesByID[obj.ID] = obj
	})
}
