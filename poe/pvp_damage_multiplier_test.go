package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestPvPDamageMultiplier(t *testing.T) {
	initGameData(t)

	carriers := 0
	for _, level := range GrantedEffectsPerLevels {
		if level.PvPDamageMultiplier != 0 {
			carriers++
		}
	}
	testza.AssertEqual(t, 5692, carriers)

	for _, id := range []string{"StaticStrikeAltX", "SupportInvertTheRules", "Fireball"} {
		want := -20
		if id == "Fireball" {
			want = -30
		}
		for level, row := range GrantedEffectByID(id).GetEffectsPerLevel() {
			testza.AssertEqual(t, want, row.PvPDamageMultiplier, id, level)
		}
	}
}
