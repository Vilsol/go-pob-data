package poe

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestMapTables(t *testing.T) {
	initGameData(t, InitializeMapTiers, InitializeMaps, InitializeWorldAreas)

	testza.AssertEqual(t, 17, len(MapTiers))
	testza.AssertEqual(t, 68, MapTiersByTier[1].Level)
	testza.AssertEqual(t, 83, MapTiersByTier[16].Level)
	testza.AssertEqual(t, 84, MapTiersByTier[17].Level)

	area := WorldAreasByID["MapWorldsElderArenaUber"]
	testza.AssertEqual(t, 84, area.AreaLevel)
	testza.AssertTrue(t, area.IsMapArea)
	bosses := area.GetBosses()
	testza.AssertEqual(t, 2, len(bosses))
	testza.AssertEqual(t, "Metadata/Monsters/AtlasBosses/TheElder/TheElderUber", bosses[0].ID)
	testza.AssertEqual(t, "Metadata/Monsters/AtlasBosses/TheShaperBossUberElder", bosses[1].ID)

	tier16 := 0
	for _, m := range Maps {
		if m.Tier == 16 {
			tier16++
			testza.AssertTrue(t, m.GetRegularWorldArea().IsMapArea)
		}
	}
	testza.AssertTrue(t, tier16 > 0)
}

// Each uber fight is a map device recipe of four fragments at area level 85 with UberBossKeyImplicit.
func TestMapDeviceRecipes(t *testing.T) {
	initGameData(t, InitializeMapTiers, InitializeMaps, InitializeWorldAreas, InitializeMapDeviceRecipes)

	areas := map[string]string{
		"UberBossShaper":      "MapWorldsShapersRealm",
		"UberBossUberElder":   "MapWorldsElderArenaUber",
		"UberBossSirus":       "AtlasExilesBoss5",
		"UberBossCortex":      "Synthesis_MapBoss",
		"UberBossMaven":       "MavenBoss",
		"UberBossEater":       "MapWorldsPrimordialBoss4",
		"UberBossExarch":      "MapWorldsPrimordialBoss3",
		"UberBossAnger":       "MapWorldsMemoryBoss1_Uber",
		"UberBossBenevolence": "MapWorldsMemoryBoss3_Uber",
		"UberBossIgnorance":   "MapWorldsMemoryBoss2_Uber",
	}
	found := 0
	for _, recipe := range MapDeviceRecipes {
		if !strings.HasPrefix(recipe.ID, "UberBoss") {
			continue
		}
		found++
		testza.AssertEqual(t, areas[recipe.ID], recipe.GetWorldArea().ID, recipe.ID)
		testza.AssertEqual(t, 85, recipe.AreaLevel, recipe.ID)
		testza.AssertEqual(t, 4, len(recipe.RecipeItems), recipe.ID)
		mods := recipe.GetMods()
		testza.AssertEqual(t, "UberBossKeyImplicit", mods[len(mods)-1].ID, recipe.ID)
	}
	testza.AssertEqual(t, len(areas), found)
}
