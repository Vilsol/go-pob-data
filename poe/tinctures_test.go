package poe

import (
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestTinctureTables(t *testing.T) {
	initGameData(t, InitializeTinctures)

	testza.AssertGreater(t, len(Tinctures), 5)
	testza.AssertTrue(t, strings.Contains(Tinctures[0].GetBaseItem().Name, "Tincture"))
}
