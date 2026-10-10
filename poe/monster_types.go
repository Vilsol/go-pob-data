package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterType struct {
	raw.MonsterType
}

var MonsterTypes []*MonsterType

var MonsterTypesByID map[string]*MonsterType

func InitializeMonsterTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterTypes", &MonsterTypes, func(count int64) {
		MonsterTypesByID = make(map[string]*MonsterType, count)
	}, assetCache, func(obj *MonsterType) {
		MonsterTypesByID[obj.ID] = obj
	})
}
