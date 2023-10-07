package extractor

import (
	"path/filepath"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func init() {
	dataPath = filepath.Join("../", "data")
}

func TestReadIndex(t *testing.T) {
	_, err := GetBundleLoader(NewWebFS(testdata.GameVersion()))
	testza.AssertNoError(t, err)
}

func TestReadSkillGems(t *testing.T) {
	loader, err := GetBundleLoader(NewWebFS(testdata.GameVersion()))
	testza.AssertNoError(t, err)

	skillGems, err := loader.Open("Data/SkillGems.dat64")
	testza.AssertNoError(t, err)

	stat, err := skillGems.Stat()
	testza.AssertNoError(t, err)

	gemBytes := make([]byte, stat.Size())
	read, err := skillGems.Read(gemBytes)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, stat.Size(), int64(read))
}

func TestLoadSchema(t *testing.T) {
	LoadSchema(testdata.ShortGameVersion())

	testza.AssertNotNil(t, schemaFile)
	testza.AssertEqual(t, int64(3), schemaFile.Version)

	schema := GetSchema("SkillGems")
	testza.AssertEqual(t, "SkillGems", schema.Name)
	testza.AssertEqual(t, 26, len(schema.Columns))
}

func TestParseDat(t *testing.T) {
	LoadParser(testdata.ShortGameVersion())

	loader, err := GetBundleLoader(NewWebFS(testdata.GameVersion()))
	testza.AssertNoError(t, err)

	testFiles := []struct {
		Name  string
		Count int
	}{
		{Name: "Data/PassiveTreeExpansionJewels.dat64", Count: 3},
		{Name: "Data/PassiveTreeExpansionSkills.dat64", Count: 55},
		{Name: "Data/PassiveTreeExpansionSpecialSkills.dat64", Count: 307},
		{Name: "Data/CostTypes.dat64", Count: 13},
		{Name: "Data/Mods.dat64", Count: 34851},
		{Name: "Data/ActiveSkills.dat64", Count: 1438},
		{Name: "Data/Essences.dat64", Count: 105},
		{Name: "Data/CraftingBenchOptions.dat64", Count: 815},
		{Name: "Data/PantheonPanelLayout.dat64", Count: 16},
		{Name: "Data/WeaponTypes.dat64", Count: 361},
		{Name: "Data/ArmourTypes.dat64", Count: 421},
		{Name: "Data/ShieldTypes.dat64", Count: 98},
		{Name: "Data/Flasks.dat64", Count: 51},
		{Name: "Data/ComponentCharges.dat64", Count: 51},
		{Name: "Data/ComponentAttributeRequirements.dat64", Count: 797},
		{Name: "Data/BaseItemTypes.dat64", Count: 10537},
		{Name: "Data/Stats.dat64", Count: 19002},
		{Name: "Data/AlternatePassiveSkills.dat64", Count: 155},
		{Name: "Data/AlternatePassiveAdditions.dat64", Count: 94},
		{Name: "Data/DefaultMonsterStats.dat64", Count: 100},
		{Name: "Data/SkillTotemVariations.dat64", Count: 250},
		{Name: "Data/MonsterVarieties.dat64", Count: 9317},
		{Name: "Data/MonsterMapDifficulty.dat64", Count: 25},
		{Name: "Data/MonsterMapBossDifficulty.dat64", Count: 90},
		{Name: "Data/GrantedEffects.dat64", Count: 10746},
		{Name: "Data/SkillTotems.dat64", Count: 21},
		{Name: "Data/GrantedEffectStatSetsPerLevel.dat64", Count: 49872},
		{Name: "Data/GrantedEffectsPerLevel.dat64", Count: 49825},
		{Name: "Data/GrantedEffectQualityStats.dat64", Count: 1579},
		{Name: "Data/SkillGems.dat64", Count: 750},
		{Name: "Data/ItemExperiencePerLevel.dat64", Count: 321},
		{Name: "Data/Tags.dat64", Count: 1069},
	}

	for _, entry := range testFiles {
		t.Run(entry.Name, func(t *testing.T) {
			data, err := loader.Open(entry.Name)
			testza.AssertNoError(t, err)

			dat, err := ParseDat(data, filepath.Base(entry.Name))
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, entry.Count, len(dat))
		})
	}
}
