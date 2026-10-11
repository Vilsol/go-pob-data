package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestMonsterResistances(t *testing.T) {
	initGameData(t, InitializeMonsterTypes, InitializeMonsterResistances)

	testza.AssertEqual(t, 50, len(MonsterResistances))
	resistances := MonsterTypesByID["BloodMonkey"].GetResistances()
	testza.AssertEqual(t, "Fire1", resistances.ID)
	testza.AssertEqual(t, 20, resistances.FireCruel)
	testza.AssertEqual(t, 40, resistances.FireMerciless)
}
