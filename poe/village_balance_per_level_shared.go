package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type VillageBalancePerLevelShared struct {
	raw.VillageBalancePerLevelShared
}

var VillageBalancePerLevels []*VillageBalancePerLevelShared

func InitializeVillageBalancePerLevels(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "VillageBalancePerLevelShared", &VillageBalancePerLevels, nil, assetCache)
}
