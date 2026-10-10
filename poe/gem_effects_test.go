package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestGemEffectTables(t *testing.T) {
	initGameData(t, InitializeGemEffects, InitializeGemTags)

	testza.AssertGreater(t, len(GemEffects), 1000)
	fireball := GemEffectsByID["Fireball"]
	testza.AssertEqual(t, "Fireball", fireball.GetGrantedEffect().ID)
	testza.AssertGreater(t, len(fireball.GetGemTags()), 0)
	testza.AssertEqual(t, "fire", GemTagsByID["fire"].ID)
}
