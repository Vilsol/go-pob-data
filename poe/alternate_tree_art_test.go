package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestAlternateTreeArtTables(t *testing.T) {
	initGameData(t, InitializeAlternateTreeArts, InitializePassiveSkillTreeNodeFrameArts, InitializePassiveSkillTreeConnectionArts)

	testza.AssertGreater(t, len(AlternateTreeArts), 10)
	art := AlternateTreeArts[7]
	testza.AssertEqual(t, "AbyssAlternateTreeNotable", art.GetNotableFrameArt().ID)
	testza.AssertEqual(t, "AbyssAlternateTreeKeystone", art.GetKeystoneFrameArt().ID)
	testza.AssertEqual(t, "AbyssAlternateTree", art.GetConnectionArt().ID)
}
