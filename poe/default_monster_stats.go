package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type DefaultMonsterStat struct {
	raw.DefaultMonsterStat
}

var DefaultMonsterStats []*DefaultMonsterStat

func InitializeDefaultMonsterStats(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "DefaultMonsterStats", &DefaultMonsterStats, nil, assetCache)
}
