package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterVariety struct {
	raw.MonsterVariety
}

var MonsterVarieties []*MonsterVariety

func InitializeMonsterVarieties(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterVarieties", &MonsterVarieties, nil, assetCache)
}
