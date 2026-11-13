package poe

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

// 3.30 moved the frame and connection art into a PassiveSkillTreeUIArt row, which AlternateTreeArt does not carry.
func TestAlternateTreeArtTables(t *testing.T) {
	initGameDataAt(t, "3.29", InitializeAlternateTreeArts, InitializePassiveSkillTreeNodeFrameArts, InitializePassiveSkillTreeConnectionArts)

	testza.AssertGreater(t, len(AlternateTreeArts), 10)
	art := AlternateTreeArts[7]
	testza.AssertEqual(t, "AbyssAlternateTreeNotable", art.GetNotableFrameArt().ID)
	testza.AssertEqual(t, "AbyssAlternateTreeKeystone", art.GetKeystoneFrameArt().ID)
	testza.AssertEqual(t, "AbyssAlternateTree", art.GetConnectionArt().ID)
}
