package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestGrantedEffectStatSets(t *testing.T) {
	initGameData(t)

	conflagration := GrantedEffectByID("Conflagration")
	testza.AssertEqual(t, 11023, conflagration.GrantedEffectStatSets)
	testza.AssertEqual(t, "Conflagration", conflagration.GetGrantedEffectStatSet().ID)

	withStatSet2 := 0
	for _, effect := range GrantedEffects {
		if len(effect.StatSet2) > 0 {
			withStatSet2++
			testza.AssertNotNil(t, effect.GetAdditionalStatSets()[0])
		}
	}
	testza.AssertEqual(t, 35, withStatSet2)
}
