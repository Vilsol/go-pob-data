package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestModColumns(t *testing.T) {
	initGameData(t, InitializeModFamilies, InitializeBuffTemplates)

	testza.AssertGreater(t, len(ModFamilies), 5000)

	byID := make(map[string]*Mod, len(Mods))
	for _, mod := range Mods {
		byID[mod.ID] = mod
	}

	ceaseless := byID["CeaselessFleshGrantedSkillUnique__1"]
	testza.AssertEqual(t, "CeaselessFleshGrantedSkill", ceaseless.GetFamilies()[0].ID)
	testza.AssertEqual(t, 283, *ceaseless.BuffTemplate, "read from BuffTemplate1")

	arcane := byID["MonsterArchnemesisArcaneEnchanted"]
	templates := arcane.GetBuffTemplates()
	testza.AssertLen(t, templates, 2)
	testza.AssertEqual(t, BuffTemplates[70], templates[0])
	testza.AssertEqual(t, BuffTemplates[372], templates[1])
}

func TestSkillGemColumns(t *testing.T) {
	initGameData(t, InitializeGemEffects, InitializeGemTags)

	fireball := BaseItemTypeByIDMap["Metadata/Items/Gems/SkillGemFireball"].SkillGem()
	testza.AssertEqual(t, 100, fireball.Int, "read from IntelligenceRequirementPercent")
	testza.AssertEqual(t, 3, fireball.GemColour)
	testza.AssertFalse(t, fireball.IsSupport)
	testza.AssertEqual(t, "Fireball", fireball.GetGemVariants()[0].ID)

	levels := fireball.GetLevelProgression()
	testza.AssertLen(t, levels, 20)
	testza.AssertEqual(t, 1, levels[0].ItemCurrentLevel)
	testza.AssertEqual(t, 20, levels[19].ItemCurrentLevel)

	supports := 0
	for _, gem := range SkillGems {
		if gem.IsSupport {
			supports++
		}
	}
	testza.AssertEqual(t, 313, supports)
}
