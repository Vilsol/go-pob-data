package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterMapBossDifficulty struct {
	raw.MonsterMapBossDifficulty
}

var MonsterMapBossDifficulties []*MonsterMapBossDifficulty

func InitializeMonsterMapBossDifficulties(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterMapBossDifficulty", &MonsterMapBossDifficulties, nil, assetCache)
}
