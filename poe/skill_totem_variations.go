package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type SkillTotemVariation struct {
	raw.SkillTotemVariation
}

var SkillTotemVariations []*SkillTotemVariation

func InitializeSkillTotemVariations(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "SkillTotemVariations", &SkillTotemVariations, nil, assetCache)
}
