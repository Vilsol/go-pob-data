package main

import (
	"path"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe2"
)

// absentFromPoE2 are exported for PoE1 but not shipped by PoE2.
var absentFromPoE2 = []string{
	"ComponentAttributeRequirements", "PassiveSkillTattoos", "PassiveSkillTattooTargetSets", "SkillTotems",
	"Tinctures", "VillageBalancePerLevelShared", "WeaponPassiveSkills", "WeaponPassiveSkillTypes",
}

func TestTablesToExport(t *testing.T) {
	if got := tablesToExport(extractor.PoE1); len(got) != len(filesToExport) || &got[0] != &filesToExport[0] {
		t.Fatal("poe1 must export filesToExport unchanged")
	}

	poe1 := map[string]bool{}
	for _, f := range filesToExport {
		poe1[path.Base(f.path)] = true
	}

	seen := map[string]bool{}
	shared := 0
	for _, f := range tablesToExport(extractor.PoE2) {
		name, ok := strings.CutPrefix(f.path, "Data/Balance/")
		if !ok || strings.Contains(name, "/") {
			t.Errorf("%s is not a PoE2 table path", f.path)
		}
		// The raw structs are PoE1 layouts; a codec encodes poe2's rows.
		if rowType, ok := poe2.Tables[name]; ok {
			if f.Encode == nil || f.Variant != loader.TupleVariant(reflect.New(rowType).Interface()) {
				t.Errorf("%s has no codec of its poe2 rows", name)
			}
		} else if f.Encode != nil || f.Variant != "" {
			t.Errorf("%s has a codec but no poe2 rows", name)
		}
		if seen[name] {
			t.Errorf("%s is listed twice", name)
		}
		seen[name] = true
		if slices.Contains(absentFromPoE2, name) {
			t.Errorf("%s is absent from PoE2", name)
		}
		if poe1[name] {
			shared++
		}
	}

	for name := range poe2.Tables {
		if !seen[name] {
			t.Errorf("poe2 has rows of %s, which is not exported", name)
		}
	}

	if want := len(filesToExport) - len(absentFromPoE2); shared != want {
		t.Errorf("poe2 exports %d of the poe1 tables, want every present one (%d)", shared, want)
	}
}
