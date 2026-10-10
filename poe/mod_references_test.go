package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestModReferenceTables(t *testing.T) {
	initGameData(t, InitializeModTypes, InitializeBuffTemplates, InitializeVirtualStatContextFlags)

	testza.AssertGreater(t, len(ModTypes), 10000)
	testza.AssertGreater(t, len(BuffTemplates), 500)
	testza.AssertGreater(t, len(VirtualStatContextFlags), 20)
	testza.AssertEqual(t, 3, len(PassiveTreeExpansionJewelSizes))
}
