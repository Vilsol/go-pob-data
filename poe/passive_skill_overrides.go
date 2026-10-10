package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillOverride struct {
	raw.PassiveSkillOverride
}

var PassiveSkillOverrides []*PassiveSkillOverride

var PassiveSkillOverridesByID map[string]*PassiveSkillOverride

func InitializePassiveSkillOverrides(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillOverrides", &PassiveSkillOverrides, func(count int64) {
		PassiveSkillOverridesByID = make(map[string]*PassiveSkillOverride, count)
	}, assetCache, func(obj *PassiveSkillOverride) {
		PassiveSkillOverridesByID[obj.ID] = obj
	})
}

func (p *PassiveSkillOverride) GetStats() []*Stat {
	out := make([]*Stat, len(p.Stats))
	for i, key := range p.Stats {
		out[i] = Stats[key]
	}

	return out
}

func (p *PassiveSkillOverride) GetType() *PassiveSkillOverrideType {
	return PassiveSkillOverrideTypes[p.Type]
}

func (p *PassiveSkillOverride) GetLimit() *PassiveOverrideLimit {
	if p.Limit == nil {
		return nil
	}

	return PassiveOverrideLimits[*p.Limit]
}

func (p *PassiveSkillOverride) GetEffect() *GrantedEffectsPerLevel {
	if p.Effect == nil {
		return nil
	}

	return GrantedEffectsPerLevels[*p.Effect]
}
