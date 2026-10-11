package poe

import (
	"context"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
)

func TestCraftingColumns(t *testing.T) {
	initGameData(t, InitializeCraftingItemClassCategories)

	t.Run("Essences", func(t *testing.T) {
		var greed *Essence
		for _, essence := range Essences {
			if BaseItemTypes[essence.BaseItemTypesKey].ID == "Metadata/Items/Currency/CurrencyEssenceGreed1" {
				greed = essence
			}
		}
		testza.AssertEqual(t, 1, greed.Level, "PoB's Tier")
		testza.AssertEqual(t, "LifeLeechPermyriadSuffixEssence1", Mods[*greed.DisplayQuiverModsKey].ID, "PoB's QuiverMod")
		testza.AssertEqual(t, []int{1}, greed.DropLevel)
		testza.AssertEqual(t, []int{1}, greed.Level2)
		testza.AssertLen(t, greed.MemoryLines, 1)
	})

	t.Run("CraftingItemClassCategories", func(t *testing.T) {
		testza.AssertGreater(t, len(CraftingItemClassCategories), 20)
		oneHand := CraftingItemClassCategoriesByID["OneHandMelee"]
		testza.AssertEqual(t, "1HandedMelee", oneHand.CategoryCode)
		testza.AssertGreater(t, len(oneHand.GetItemClasses()), 5)

		var bench *CraftingBenchOption
		for _, option := range CraftingBenchOptions {
			if option.AddMod != nil && option.GetMod().ID == "StrDexIntMasterMapQuantity1" {
				bench = option
			}
		}
		testza.AssertLen(t, bench.GetItemCategories(), 4)
	})

	t.Run("BaseItemTypes", func(t *testing.T) {
		enchanted := 0
		for _, base := range BaseItemTypes {
			if len(base.EnchantMods) > 0 {
				enchanted++
			}
		}
		testza.AssertEqual(t, 43, enchanted, "PoB's cannotBeAnointed bases")
		testza.AssertEqual(t, "TalismanEnchantAreaOfEffect", BaseItemTypeByNameMap["Chieftain Talisman"].GetEnchantMods()[0].ID)
	})

	t.Run("Mods", func(t *testing.T) {
		flagged := 0
		for _, mod := range Mods {
			if len(mod.AuraFlags) > 0 {
				flagged++
			}
		}
		testza.AssertEqual(t, 4, flagged)
	})

	t.Run("ActiveSkills", func(t *testing.T) {
		var slam *ActiveSkill
		for _, skill := range ActiveSkills {
			if skill.ID == "ground_slam" {
				slam = skill
			}
		}
		testza.AssertEqual(t, "ground_slam_angle_+%", slam.GetSecondarySkillSpecificStats()[0].ID)
		testza.AssertGreater(t, len(slam.GetSkillSpecificStats()), 0)
	})
}

func TestItemTemplateFiles(t *testing.T) {
	testdata.ServeCommittedData(t)
	version := testdata.LatestDataVersion()

	m, err := loader.LoadManifest(context.Background(), version)
	testza.AssertNoError(t, err)
	testza.AssertTrue(t, m.HasFile("Metadata/Items/Item.it"), "the root every base extends")

	amulet, err := loader.LoadFile(context.Background(), version, "Metadata/Items/Amulets/AbstractAmulet.it", FakeCache{})
	testza.AssertNoError(t, err)
	testza.AssertContains(t, string(amulet), `extends "Metadata/Items/Equipment"`)
	testza.AssertContains(t, string(amulet), `tag = "amulet"`)
}
