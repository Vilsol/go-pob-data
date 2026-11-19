package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterPack struct {
	raw.MonsterPack
}

var MonsterPacks []*MonsterPack

var MonsterPacksByID map[string]*MonsterPack

func InitializeMonsterPacks(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterPacks", &MonsterPacks, func(count int64) {
		MonsterPacksByID = make(map[string]*MonsterPack, count)
	}, assetCache, func(obj *MonsterPack) {
		MonsterPacksByID[obj.ID] = obj
	})
}
