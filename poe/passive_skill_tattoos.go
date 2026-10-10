package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillTattoo struct {
	raw.PassiveSkillTattoo
}

var PassiveSkillTattoos []*PassiveSkillTattoo

func InitializePassiveSkillTattoos(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillTattoos", &PassiveSkillTattoos, nil, assetCache)
}

func (p *PassiveSkillTattoo) GetTattoo() *BaseItemType {
	return BaseItemTypes[p.Tattoo]
}

func (p *PassiveSkillTattoo) GetOverride() *PassiveSkillOverride {
	return PassiveSkillOverrides[p.Override]
}

func (p *PassiveSkillTattoo) GetSet() *PassiveSkillTattooTargetSet {
	return PassiveSkillTattooTargetSets[p.Set]
}

func (p *PassiveSkillTattoo) GetOverrideType() *PassiveSkillOverrideType {
	if p.OverrideType == nil {
		return nil
	}

	return PassiveSkillOverrideTypes[*p.OverrideType]
}
