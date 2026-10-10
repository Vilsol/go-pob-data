package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillTattooTargetSet struct {
	raw.PassiveSkillTattooTargetSet
}

var PassiveSkillTattooTargetSets []*PassiveSkillTattooTargetSet

var PassiveSkillTattooTargetSetsByID map[string]*PassiveSkillTattooTargetSet

func InitializePassiveSkillTattooTargetSets(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillTattooTargetSets", &PassiveSkillTattooTargetSets, func(count int64) {
		PassiveSkillTattooTargetSetsByID = make(map[string]*PassiveSkillTattooTargetSet, count)
	}, assetCache, func(obj *PassiveSkillTattooTargetSet) {
		PassiveSkillTattooTargetSetsByID[obj.ID] = obj
	})
}
