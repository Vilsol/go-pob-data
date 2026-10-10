package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type WeaponPassiveSkillType struct {
	raw.WeaponPassiveSkillType
}

var WeaponPassiveSkillTypes []*WeaponPassiveSkillType

var WeaponPassiveSkillTypesByID map[string]*WeaponPassiveSkillType

func InitializeWeaponPassiveSkillTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "WeaponPassiveSkillTypes", &WeaponPassiveSkillTypes, func(count int64) {
		WeaponPassiveSkillTypesByID = make(map[string]*WeaponPassiveSkillType, count)
	}, assetCache, func(obj *WeaponPassiveSkillType) {
		WeaponPassiveSkillTypesByID[obj.ID] = obj
	})
}
