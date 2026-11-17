package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

// The tables behind 3.30's stat description formatter functions (PoB's statdesc.lua
// getDescriptionFormatterNames), and the Mods column that shows a mod's granted effects.
func TestStatDescriptionFormatterTables(t *testing.T) {
	initGameData(t, InitializeClientStrings, InitializePassiveJewelRadii, InitializeFishingLureTypes, InitializeFishingHookTypes,
		InitializeShaperApparitionAbilities, InitializeMinionDoublingStatTypes, InitializeAtlasExiles, InitializeElderGuardians)

	testza.AssertEqual(t, 5, len(PassiveJewelRadii))
	testza.AssertEqual(t, "Massive", ClientStrings[*PassiveJewelRadii[4].ClientString].Text)
	testza.AssertEqual(t, "Kulemak's Corpsebait", FishingLureTypes[6].Name)
	testza.AssertEqual(t, "Eternal Iron Hook", FishingHookTypes[3].Name)
	testza.AssertEqual(t, "Shaper's Despair", ShaperApparitionAbilities[2].Name)
	testza.AssertEqual(t, "Holy Armaments", MinionDoublingStatTypes[15].Name)
	testza.AssertEqual(t, "Drox's Citadel", AtlasExiles[3].CitadelName)
	testza.AssertEqual(t, "The Purifier", ElderGuardians[3].Name)

	show := map[string]bool{}
	for _, m := range Mods {
		show[m.ID] = m.ShowGrantedEffects
	}
	testza.AssertTrue(t, show["VisceraModBasiliskHigh"])
	testza.AssertFalse(t, show["VisceraModArakaaliHigh"])
}
