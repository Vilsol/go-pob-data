package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveSkillOverrideType struct {
	raw.PassiveSkillOverrideType
}

var PassiveSkillOverrideTypes []*PassiveSkillOverrideType

var PassiveSkillOverrideTypesByID map[string]*PassiveSkillOverrideType

func InitializePassiveSkillOverrideTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveSkillOverrideTypes", &PassiveSkillOverrideTypes, func(count int64) {
		PassiveSkillOverrideTypesByID = make(map[string]*PassiveSkillOverrideType, count)
	}, assetCache, func(obj *PassiveSkillOverrideType) {
		PassiveSkillOverrideTypesByID[obj.ID] = obj
	})
}

func (p *PassiveSkillOverrideType) GetCounterStat() *Stat {
	if p.CounterStat == nil {
		return nil
	}

	return Stats[*p.CounterStat]
}
