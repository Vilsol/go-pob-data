package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ActiveSkill struct {
	raw.ActiveSkill
}

var ActiveSkills []*ActiveSkill

var ActiveSkillTypesBySkillID map[string]*ActiveSkill

func InitializeActiveSkills(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ActiveSkills", &ActiveSkills, func(count int64) {
		ActiveSkillTypesBySkillID = make(map[string]*ActiveSkill, count)
	}, assetCache, func(obj *ActiveSkill) {
		ActiveSkillTypesBySkillID[obj.SkillID] = obj
	})
}

func (g *ActiveSkill) GetActiveSkillTypes() []*ActiveSkillType {
	if g.ActiveSkillTypes == nil {
		return nil
	}

	out := make([]*ActiveSkillType, len(g.ActiveSkillTypes))
	for i, skillType := range g.ActiveSkillTypes {
		out[i] = ActiveSkillTypes[skillType]
	}

	return out
}

func (g *ActiveSkill) GetWeaponRestrictions() []*ItemClass {
	if g.WeaponRestrictionItemClassesKeys == nil {
		return nil
	}

	out := make([]*ItemClass, len(g.WeaponRestrictionItemClassesKeys))
	for i, skillType := range g.WeaponRestrictionItemClassesKeys {
		out[i] = ItemClasses[skillType]
	}

	return out
}

func (g *ActiveSkill) GetActiveSkillBaseFlagsAndTypes() (map[SkillFlag]bool, map[SkillType]bool) {
	// TODO Cache
	flags := make(map[SkillFlag]bool)
	types := make(map[SkillType]bool)
	for _, skillTypeRaw := range g.GetActiveSkillTypes() {
		skillType := SkillType(skillTypeRaw.ID)
		types[skillType] = true

		switch skillType {
		case SkillTypeBrand:
			flags[SkillFlagBrand] = true
		case SkillTypeHex:
			flags[SkillFlagHex] = true
			flags[SkillFlagCurse] = true
		case SkillTypeAppliesCurse:
			flags[SkillFlagCurse] = true
		case SkillTypeAttack:
			flags[SkillFlagAttack] = true
			flags[SkillFlagHit] = true
		case SkillTypeProjectile:
			flags[SkillFlagProjectile] = true
			flags[SkillFlagHit] = true
		case SkillTypeTrapped:
			flags[SkillFlagTrap] = true
		case SkillTypeTrappable:
			flags[SkillFlagTrap] = true
		case SkillTypeRemoteMined:
			flags[SkillFlagMine] = true
		case SkillTypeSummonsTotem:
			flags[SkillFlagTotem] = true
		case SkillTypeSpell:
			flags[SkillFlagSpell] = true
		case SkillTypeAreaSpell:
			flags[SkillFlagSpell] = true
			flags[SkillFlagArea] = true
		case SkillTypeMelee:
			flags[SkillFlagMelee] = true
		case SkillTypeMeleeSingleTarget:
			flags[SkillFlagMelee] = true
		case SkillTypeChains:
			flags[SkillFlagChaining] = true
		case SkillTypeArea:
			flags[SkillFlagArea] = true
		case SkillTypeDamage:
			flags[SkillFlagHit] = true
			// TODO SkillFlagCast
			// case data.SkillType...:
			//	 flags[SkillFlagCast] = true
		}
	}
	return flags, types
}
