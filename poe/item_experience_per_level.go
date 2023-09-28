package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ItemExperiencePerLevel struct {
	raw.ItemExperiencePerLevel
}

var ItemExperiencePerLevels []*ItemExperiencePerLevel

func InitializeItemExperiencePerLevels(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ItemExperiencePerLevel", &ItemExperiencePerLevels, nil, assetCache)
}
