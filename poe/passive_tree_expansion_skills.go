package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveTreeExpansionSkill struct {
	raw.PassiveTreeExpansionSkill
}

var PassiveTreeExpansionSkills []*PassiveTreeExpansionSkill

func InitializePassiveTreeExpansionSkills(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveTreeExpansionSkills", &PassiveTreeExpansionSkills, nil, assetCache)
}
