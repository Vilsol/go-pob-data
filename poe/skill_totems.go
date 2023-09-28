package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type SkillTotem struct {
	raw.SkillTotem
}

var SkillTotems []*SkillTotem

func InitializeSkillTotems(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "SkillTotems", &SkillTotems, nil, assetCache)
}
