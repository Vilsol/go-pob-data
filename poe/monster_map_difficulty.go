package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterMapDifficulty struct {
	raw.MonsterMapDifficulty
}

var MonsterMapDifficulties []*MonsterMapDifficulty

func InitializeMonsterMapDifficulties(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterMapDifficulty", &MonsterMapDifficulties, nil, assetCache)
}
