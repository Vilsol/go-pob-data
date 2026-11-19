package poe

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/raw"
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

func moreLifeAt(rows []*raw.LifeScalingPerLevel, level int) (int, bool) {
	for _, row := range rows {
		if row.Level == level {
			return row.MoreLife, true
		}
	}
	return 0, false
}

func TestAdditionalLifeScaling(t *testing.T) {
	initGameData(t, InitializeAdditionalLifeScalings)

	primordial := AdditionalLifeScalingsByIntID[8]
	testza.AssertEqual(t, "PrimordialBossLifeScalingPerLevel", primordial.ID)
	for level, want := range map[int]int{79: 0, 83: 1550, 84: 4500, 85: 4500} {
		got, ok := moreLifeAt(primordial.PerLevel, level)
		testza.AssertTrue(t, ok, level)
		testza.AssertEqual(t, want, got, level)
	}

	uber := AdditionalLifeScalingsByIntID[21]
	testza.AssertEqual(t, "AtlasUberMonsterLifeScalingPerLevel", uber.ID)
	for level, want := range map[int]int{68: 140, 83: 275, 84: 275} {
		got, _ := moreLifeAt(uber.PerLevel, level)
		testza.AssertEqual(t, want, got, level)
	}

	// Its columns are Level and Life, not MoreLife: not published.
	testza.AssertEqual(t, "DeepwaterLifeScalingPerLevel", AdditionalLifeScalingsByIntID[26].ID)
	testza.AssertNil(t, AdditionalLifeScalingsByIntID[26].PerLevel)
}

func TestBuffDefinitions(t *testing.T) {
	initGameData(t, InitializeBuffDefinitions)

	buffStats := BuffDefinitionsByID["uber_elder_immune_buff"].GetStats()
	stats := make([]string, len(buffStats))
	for i, stat := range buffStats {
		stats[i] = stat.ID
	}
	testza.AssertEqual(t, []string{
		"base_cannot_be_damaged", "base_cannot_be_stunned", "immune_to_auras_from_other_entities", "immune_to_curses",
		"untargetable_by_monster_ai", "cannot_die",
	}, stats)
}
