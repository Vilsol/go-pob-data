package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestUniqueTables(t *testing.T) {
	initGameData(t, InitializeFlavourTexts, InitializeUniqueStashLayouts, InitializeUniqueStashTypes, InitializeWords, InitializeItemVisualIdentities)

	t.Run("FlavourText", func(t *testing.T) {
		testza.AssertGreater(t, len(FlavourTexts), 2000)
		testza.AssertEqual(t, "Forged from thousands of unclaimed Beta keys", FlavourTextsByID["KeybladeText"].Text)
	})

	t.Run("UniqueStashLayout", func(t *testing.T) {
		testza.AssertGreater(t, len(UniqueStashLayouts), 1000)
		layout := UniqueStashLayouts[0]
		testza.AssertEqual(t, "Kaom's Primacy", layout.GetWord().Text)
		testza.AssertEqual(t, "Axe", layout.GetUniqueStashType().ID)
		testza.AssertNotNil(t, layout.GetItemVisualIdentity())
		testza.AssertGreater(t, len(Words), 3000)
		testza.AssertGreater(t, len(UniqueStashTypes), 20)
		testza.AssertGreater(t, len(ItemVisualIdentities), 10000)
	})
}
