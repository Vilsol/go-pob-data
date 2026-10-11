package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestMonsterStatColumns(t *testing.T) {
	initGameData(t)

	level1 := DefaultMonsterStats[0]
	testza.AssertEqual(t, "1", level1.DisplayLevel)
	testza.AssertEqual(t, 22, level1.MonsterLife)
	testza.AssertEqual(t, 15, level1.MinionLife)
	testza.AssertEqual(t, 5.62, level1.MinionDamage)
	testza.AssertEqual(t, 22, level1.AilmentThreshold)

	for _, row := range MonsterMapDifficulties {
		if row.MapLevel == 70 {
			testza.AssertEqual(t, 13, row.LifePercentIncrease)
			testza.AssertNotNil(t, row.GetStats()[0])
		}
	}
	for _, row := range MonsterMapBossDifficulties {
		if row.MapLevel == 70 {
			testza.AssertEqual(t, 216, row.BossLifePercentIncrease)
			testza.AssertEqual(t, -42, row.BossAilmentPercentDecrease)
			testza.AssertLen(t, row.GetStats(), 5)
		}
	}
}
