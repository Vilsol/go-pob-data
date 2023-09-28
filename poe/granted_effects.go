package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GrantedEffect struct {
	raw.GrantedEffect
}

var GrantedEffects []*GrantedEffect

var grantedEffectsByIDMap map[string]*GrantedEffect

func InitializeGrantedEffects(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GrantedEffects", &GrantedEffects, func(count int64) {
		grantedEffectsByIDMap = make(map[string]*GrantedEffect, count)
	}, assetCache, func(obj *GrantedEffect) {
		grantedEffectsByIDMap[obj.ID] = obj
	})
}

func GrantedEffectByID(id string) *GrantedEffect {
	return grantedEffectsByIDMap[id]
}

func (g *GrantedEffect) GetActiveSkill() *ActiveSkill {
	if g.ActiveSkill == nil {
		return nil
	}

	return ActiveSkills[*g.ActiveSkill]
}

func (g *GrantedEffect) GetEffectsPerLevel() map[int]*GrantedEffectsPerLevel {
	return grantedEffectsPerLevelsByIDMap[g.Key]
}

func (g *GrantedEffect) GetEffectStatSetsPerLevel() map[int]*GrantedEffectStatSetsPerLevel {
	return grantedEffectStatSetsPerLevelsByIDMap[g.Key]
}

func (g *GrantedEffect) GetEffectQualityStats() map[int]*GrantedEffectQualityStat {
	return grantedEffectQualityStatsByIDMap[g.Key]
}

func (g *GrantedEffect) GetSkillGem() *SkillGem {
	return skillGemsByGrantedEffect[g.Key]
}

func (g *GrantedEffect) HasGlobalEffect() bool {
	// TODO HasGlobalEffect
	return false
}

func (g *GrantedEffect) Levels() map[int]*GrantedEffectsPerLevel {
	return grantedEffectsPerLevelsByIDMap[g.Key]
}

func (g *GrantedEffect) GetGrantedEffectStatSet() *GrantedEffectStatSet {
	return GrantedEffectStatSets[g.GrantedEffectStatSets]
}

func (g *GrantedEffect) GetSupportTypes() []*ActiveSkillType {
	if g.SupportTypes == nil {
		return nil
	}
	out := make([]*ActiveSkillType, len(g.SupportTypes))
	for i, supportType := range g.SupportTypes {
		out[i] = ActiveSkillTypes[supportType]
	}
	return out
}

func (g *GrantedEffect) GetExcludeTypes() []*ActiveSkillType {
	if g.ExcludeTypes == nil {
		return nil
	}
	out := make([]*ActiveSkillType, len(g.ExcludeTypes))
	for i, supportType := range g.ExcludeTypes {
		out[i] = ActiveSkillTypes[supportType]
	}
	return out
}
