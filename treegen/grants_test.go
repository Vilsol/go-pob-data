package treegen

import (
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestGrantLine(t *testing.T) {
	skill := 0
	g := &grants{
		levels:        []grantedEffectLevel{{GrantedEffect: 0, Level: 20}, {GrantedEffect: 1, Level: 20}},
		effects:       []grantedEffect{{ID: "Barkskin", ActiveSkill: &skill}, {ID: "PrimalAegis", ActiveSkill: &skill, CannotBeSupported: true}},
		activeSkills:  []activeSkill{{DisplayedName: "Some Skill"}},
		text:          "Grants {0}",
		textWithLevel: "Grants Level {1} {0}",
	}

	line, err := g.line(0)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, "Grants Level 20 [Barkskin::20|Some Skill]", line)

	line, err = g.line(1)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, "Grants [PrimalAegis::20|Some Skill]", line)

	d := &gameData{}
	lines, err := d.grantLines(nil)
	testza.AssertNoError(t, err)
	testza.AssertNil(t, lines)
}

func TestLoadGrantsWithoutClientStrings(t *testing.T) {
	l := &tableLoader{}
	testza.AssertNil(t, loadGrants(l, []clientString{{ID: grantsPassivePoint, Text: "Grants {0} Passive Skill Point"}}))
	testza.AssertNoError(t, l.err)

	loadGrants(l, []clientString{{ID: grantsEffect, Text: "Grants {0}"}})
	testza.AssertNotNil(t, l.err)
}
