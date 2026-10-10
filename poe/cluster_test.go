package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestClusterTables(t *testing.T) {
	initGameData(t, InitializePassiveJewelSlots, InitializePassiveTreeExpansionJewelSizes)

	testza.AssertGreater(t, len(PassiveJewelSlots), 50)
	sizes := 0
	for _, slot := range PassiveJewelSlots {
		if slot.GetClusterJewelSize() != nil {
			sizes++
		}
	}
	testza.AssertGreater(t, sizes, 0)
}
