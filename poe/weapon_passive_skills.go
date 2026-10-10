package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type WeaponPassiveSkill struct {
	raw.WeaponPassiveSkill
}

var WeaponPassiveSkills []*WeaponPassiveSkill

// WeaponPassiveSkillsByID holds every tier of a skill, in key order: tiers share their Id.
var WeaponPassiveSkillsByID map[string][]*WeaponPassiveSkill

func InitializeWeaponPassiveSkills(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "WeaponPassiveSkills", &WeaponPassiveSkills, func(_ int64) {
		WeaponPassiveSkillsByID = make(map[string][]*WeaponPassiveSkill)
	}, assetCache, func(obj *WeaponPassiveSkill) {
		WeaponPassiveSkillsByID[obj.ID] = append(WeaponPassiveSkillsByID[obj.ID], obj)
	})
}

func (w *WeaponPassiveSkill) GetMod() *Mod {
	return Mods[w.Mod]
}

func (w *WeaponPassiveSkill) GetType() *WeaponPassiveSkillType {
	return WeaponPassiveSkillTypes[w.Type]
}
