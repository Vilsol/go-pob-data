package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type EffectivenessCostConstant struct {
	raw.EffectivenessCostConstant
}

var EffectivenessCostConstants []*EffectivenessCostConstant

var EffectivenessCostConstantsByID map[string]*EffectivenessCostConstant

func InitializeEffectivenessCostConstants(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "EffectivenessCostConstants", &EffectivenessCostConstants, func(count int64) {
		EffectivenessCostConstantsByID = make(map[string]*EffectivenessCostConstant, count)
	}, assetCache, func(obj *EffectivenessCostConstant) {
		EffectivenessCostConstantsByID[obj.ID] = obj
	})
}
