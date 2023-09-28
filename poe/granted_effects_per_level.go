package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GrantedEffectsPerLevel struct {
	raw.GrantedEffectsPerLevel
}

var GrantedEffectsPerLevels []*GrantedEffectsPerLevel

var grantedEffectsPerLevelsByIDMap map[int]map[int]*GrantedEffectsPerLevel

func InitializeGrantedEffectsPerLevels(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GrantedEffectsPerLevel", &GrantedEffectsPerLevels, func(count int64) {
		grantedEffectsPerLevelsByIDMap = make(map[int]map[int]*GrantedEffectsPerLevel, count)
	}, assetCache, func(obj *GrantedEffectsPerLevel) {
		if _, ok := grantedEffectsPerLevelsByIDMap[obj.GrantedEffect]; !ok {
			grantedEffectsPerLevelsByIDMap[obj.GrantedEffect] = make(map[int]*GrantedEffectsPerLevel)
		}

		grantedEffectsPerLevelsByIDMap[obj.GrantedEffect][obj.Level] = obj
	})
}

func (l *GrantedEffectsPerLevel) GetCostTypes() []*CostType {
	if l.CostTypes == nil {
		return nil
	}

	out := make([]*CostType, len(l.CostTypes))
	for i, costType := range l.CostTypes {
		out[i] = CostTypes[costType]
	}
	return out
}
