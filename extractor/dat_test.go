package extractor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func init() {
	dataPath = filepath.Join("../", "data")
}

// The patch CDN only serves the live patch, so these tests follow it and skip until its schema is committed.
func requireLiveSchema(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dataPath, testdata.ShortGameVersion(), "schema.min.json")); err != nil {
		t.Skipf("no committed schema for live version %s", testdata.ShortGameVersion())
	}
}

func datPath(name string) string {
	return "data/" + strings.ToLower(name) + ".datc64"
}

func TestReadIndex(t *testing.T) {
	_, err := GetBundleLoader(NewWebFS(PoE1, testdata.GameVersion()), true)
	testza.AssertNoError(t, err)
}

func TestReadSkillGems(t *testing.T) {
	loader, err := GetBundleLoader(NewWebFS(PoE1, testdata.GameVersion()), true)
	testza.AssertNoError(t, err)

	skillGems, err := loader.Open(datPath("SkillGems"))
	testza.AssertNoError(t, err)

	stat, err := skillGems.Stat()
	testza.AssertNoError(t, err)

	gemBytes := make([]byte, stat.Size())
	read, err := skillGems.Read(gemBytes)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, stat.Size(), int64(read))
}

func TestLoadSchema(t *testing.T) {
	requireLiveSchema(t)
	LoadSchema(PoE1, testdata.ShortGameVersion())

	testza.AssertNotNil(t, schemaFile)

	schema := GetSchema("SkillGems")
	testza.AssertEqual(t, "SkillGems", schema.Name)
	testza.AssertGreater(t, len(schema.Columns), 0)
}

func TestParseDat(t *testing.T) {
	requireLiveSchema(t)
	LoadParser(PoE1, testdata.ShortGameVersion())

	loader, err := GetBundleLoader(NewWebFS(PoE1, testdata.GameVersion()), true)
	testza.AssertNoError(t, err)

	tables := []string{
		"PassiveTreeExpansionJewels",
		"PassiveTreeExpansionSkills",
		"PassiveTreeExpansionSpecialSkills",
		"CostTypes",
		"Mods",
		"ActiveSkills",
		"Essences",
		"CraftingBenchOptions",
		"PantheonPanelLayout",
		"WeaponTypes",
		"ArmourTypes",
		"ShieldTypes",
		"Flasks",
		"ComponentCharges",
		"ComponentAttributeRequirements",
		"BaseItemTypes",
		"Stats",
		"AlternatePassiveSkills",
		"AlternatePassiveAdditions",
		"DefaultMonsterStats",
		"SkillTotemVariations",
		"MonsterVarieties",
		"MonsterMapDifficulty",
		"MonsterMapBossDifficulty",
		"GrantedEffects",
		"GrantedEffectStatSetsPerLevel",
		"GrantedEffectsPerLevel",
		"GrantedEffectQualityStats",
		"SkillGems",
		"ItemExperiencePerLevel",
		"Tags",
	}

	for _, name := range tables {
		t.Run(name, func(t *testing.T) {
			data, err := loader.Open(datPath(name))
			testza.AssertNoError(t, err)

			dat, err := ParseDat(data, name+".datc64")
			testza.AssertNoError(t, err)
			testza.AssertGreater(t, len(dat), 0)
		})
	}
}
