package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GrantedEffectQualityStat struct {
	raw.GrantedEffectQualityStat
}

var GrantedEffectQualityStats []*GrantedEffectQualityStat

var grantedEffectQualityStatsByIDMap map[int]map[int]*GrantedEffectQualityStat

func InitializeGrantedEffectQualityStats(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GrantedEffectQualityStats", &GrantedEffectQualityStats, func(count int64) {
		grantedEffectQualityStatsByIDMap = make(map[int]map[int]*GrantedEffectQualityStat, count)
	}, assetCache, func(obj *GrantedEffectQualityStat) {
		if _, ok := grantedEffectQualityStatsByIDMap[obj.GrantedEffectsKey]; !ok {
			grantedEffectQualityStatsByIDMap[obj.GrantedEffectsKey] = make(map[int]*GrantedEffectQualityStat)
		}

		grantedEffectQualityStatsByIDMap[obj.GrantedEffectsKey][obj.SetID] = obj
	})
}

func (s *GrantedEffectQualityStat) GetStats() []*Stat {
	if s.StatsKeys == nil {
		return nil
	}

	out := make([]*Stat, len(s.StatsKeys))
	for i, key := range s.StatsKeys {
		out[i] = Stats[key]
	}
	return out
}
