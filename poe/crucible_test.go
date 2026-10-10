package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestCrucibleTables(t *testing.T) {
	initGameData(t, InitializeWeaponPassiveSkills, InitializeWeaponPassiveSkillTypes)

	testza.AssertGreater(t, len(WeaponPassiveSkills), 2000)
	tiers := WeaponPassiveSkillsByID["AddedPhysicalHighReducedAttackSpeed"]
	testza.AssertGreater(t, len(tiers), 1)
	skill := tiers[0]
	testza.AssertEqual(t, 1, skill.Tier)
	testza.AssertEqual(t, "WeaponTreeAddedPhysicalHighReducedAttackSpeed1", skill.GetMod().ID)
	testza.AssertEqual(t, "Regular", skill.GetType().ID)
}
