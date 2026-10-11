package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterResistance struct {
	raw.MonsterResistance
}

var MonsterResistances []*MonsterResistance

var MonsterResistancesByID map[string]*MonsterResistance

func InitializeMonsterResistances(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterResistances", &MonsterResistances, func(count int64) {
		MonsterResistancesByID = make(map[string]*MonsterResistance, count)
	}, assetCache, func(obj *MonsterResistance) {
		MonsterResistancesByID[obj.ID] = obj
	})
}
