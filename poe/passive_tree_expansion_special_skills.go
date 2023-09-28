package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveTreeExpansionSpecialSkill struct {
	raw.PassiveTreeExpansionSpecialSkill
}

var PassiveTreeExpansionSpecialSkills []*PassiveTreeExpansionSpecialSkill

func InitializePassiveTreeExpansionSpecialSkills(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveTreeExpansionSpecialSkills", &PassiveTreeExpansionSpecialSkills, nil, assetCache)
}
