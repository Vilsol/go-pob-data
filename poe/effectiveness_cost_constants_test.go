package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestEffectivenessCostConstants(t *testing.T) {
	initGameData(t, InitializeEffectivenessCostConstants)

	testza.AssertLen(t, EffectivenessCostConstants, 3)
	testza.AssertEqual(t, 1.0, EffectivenessCostConstantsByID["DamageValue"].Multiplier)
	testza.AssertEqual(t, 0.06, EffectivenessCostConstantsByID["DamageValuePerMinute"].Multiplier)
	testza.AssertEqual(t, 0.0, EffectivenessCostConstantsByID["Override"].Multiplier)

	resolved := 0
	for _, level := range GrantedEffectStatSetsPerLevels {
		for _, base := range level.GetInterpolationBases() {
			testza.AssertNotNil(t, base)
			resolved++
		}
	}
	testza.AssertGreater(t, resolved, 0)
}
