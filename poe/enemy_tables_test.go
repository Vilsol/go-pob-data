package poe

import (
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
