package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestItemArtPaths(t *testing.T) {
	initGameData(t, InitializeUniqueStashLayouts, InitializeWords, InitializeItemVisualIdentities)

	t.Run("BaseItemType", func(t *testing.T) {
		for name, art := range map[string]string{
			"Vaal Regalia":              "Art/2DItems/Armours/BodyArmours/BodyStrDexInt1C",
			"Divine Life Flask":         "Art/2DItems/Flasks/lifeflask11",
			"Crimson Jewel":             "Art/2DItems/Jewels/basicstr",
			"Fireball":                  "Art/2DItems/Gems/Fireball",
			"Added Fire Damage Support": "Art/2DItems/Gems/Support/AddedFireDamage",
		} {
			testza.AssertEqual(t, art, BaseItemTypeByNameMap[name].GetItemVisualIdentity().ArtPath(), name)
		}
	})

	t.Run("Unique", func(t *testing.T) {
		for name, art := range map[string]string{
			// Headhunter and Tabula Rasa also have an alternate art row
			"Headhunter":   "Art/2DItems/Belts/Headhunter",
			"Tabula Rasa":  "Art/2DItems/Armours/BodyArmours/TabulaRasa",
			"Kaom's Heart": "Art/2DItems/Armours/BodyArmours/BodyStr3CUnique",
			"Mageblood":    "Art/2DItems/Belts/InjectorBelt",
			// Its word also exists outside the unique word list
			"Inertia":    "Art/2DItems/Jewels/DEXtoSTR",
			"Impresence": "Art/2DItems/Amulets/ElderFire",
		} {
			layout := UniqueStashLayoutByName(name)
			testza.AssertNotNil(t, layout, name)
			testza.AssertEqual(t, art, layout.GetItemVisualIdentity().ArtPath(), name)
		}
		testza.AssertNil(t, UniqueStashLayoutByName("Leather Belt"))
		testza.AssertNil(t, UniqueStashLayoutByName("Not A Unique"))
	})
}
