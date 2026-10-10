package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestConstantTables(t *testing.T) {
	initGameData(t, InitializeMonsterTypes, InitializeGameConstants, InitializeVillageBalancePerLevels)

	t.Run("MonsterTypes", func(t *testing.T) {
		testza.AssertGreater(t, len(MonsterTypes), 3000)
		testza.AssertEqual(t, 15, MonsterTypesByID["BloodMonkey"].Evasion)
	})

	t.Run("GameConstants", func(t *testing.T) {
		testza.AssertGreater(t, len(GameConstants), 30)
		constant := GameConstantsByID["MonsterDamageReductionImprovement"]
		testza.AssertEqual(t, 670, constant.Value)
		testza.AssertEqual(t, 100, constant.Divisor)
	})

	t.Run("VillageBalancePerLevelShared", func(t *testing.T) {
		testza.AssertEqual(t, 100, len(VillageBalancePerLevels))
		testza.AssertEqual(t, 4, VillageBalancePerLevels[0].GoldRespecCost)
	})
}
