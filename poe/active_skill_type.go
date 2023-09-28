package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ActiveSkillType struct {
	raw.ActiveSkillType
}

var (
	ActiveSkillTypes     []*ActiveSkillType
	ActiveSkillTypesByID map[string]*ActiveSkillType
)

func InitializeActiveSkillTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ActiveSkillType", &ActiveSkillTypes, func(count int64) {
		ActiveSkillTypesByID = make(map[string]*ActiveSkillType, count)
	}, assetCache, func(obj *ActiveSkillType) {
		ActiveSkillTypesByID[obj.ID] = obj
	})
}
