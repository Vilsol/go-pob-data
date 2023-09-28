package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type SkillGem struct {
	raw.SkillGem
}

var SkillGems []*SkillGem

var (
	skillGemsByBaseItemTypeMap map[int]*SkillGem
	skillGemVaalBase           map[int]*SkillGem
	skillGemsByGrantedEffect   map[int]*SkillGem
)

func InitializeSkillGems(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "SkillGems", &SkillGems, func(count int64) {
		skillGemsByBaseItemTypeMap = make(map[int]*SkillGem, count)
		skillGemVaalBase = make(map[int]*SkillGem, count)
		skillGemsByGrantedEffect = make(map[int]*SkillGem, count)
	}, assetCache, func(obj *SkillGem) {
		skillGemsByBaseItemTypeMap[obj.BaseItemType] = obj
		skillGemsByGrantedEffect[obj.GrantedEffect] = obj
		if obj.VaalGem != nil {
			skillGemVaalBase[*obj.VaalGem] = obj
		}
	})
}

func (s *SkillGem) GetGrantedEffect() *GrantedEffect {
	return GrantedEffects[s.GrantedEffect]
}

func (s *SkillGem) GetSecondaryGrantedEffect() *GrantedEffect {
	if s.SecondaryGrantedEffect == nil {
		return nil
	}

	return GrantedEffects[*s.SecondaryGrantedEffect]
}

func (s *SkillGem) GetGrantedEffects() []*GrantedEffect {
	out := make([]*GrantedEffect, 1)
	out[0] = s.GetGrantedEffect()

	secondary := s.GetSecondaryGrantedEffect()
	if secondary != nil {
		out = append(out, secondary)
	}

	return out
}

func (s *SkillGem) GetTags() map[raw.TagName]*Tag {
	out := make(map[raw.TagName]*Tag, len(s.Tags))
	for _, tag := range s.Tags {
		t := Tags[tag]
		out[t.Name] = t
	}
	return out
}

func (s *SkillGem) DefaultLevel() int {
	levels := s.GetGrantedEffect().Levels()
	if len(levels) > 20 {
		return len(levels) - 21
	}
	// TODO Awakened gem default level?
	return 1
}

func (s *SkillGem) GetBaseItemType() *BaseItemType {
	return BaseItemTypes[s.BaseItemType]
}

func (s *SkillGem) GetNonVaal() *SkillGem {
	return skillGemVaalBase[s.BaseItemType]
}
