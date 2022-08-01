package extractor

import (
	"path/filepath"
	"testing"

	"github.com/MarvinJWendt/testza"
)

func TestReadIndex(t *testing.T) {
	_, err := GetBundleLoader(NewWebFS())
	testza.AssertNoError(t, err)
}

func TestReadSkillGems(t *testing.T) {
	loader, err := GetBundleLoader(NewWebFS())
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
	LoadSchema()

	testza.AssertNotNil(t, schemaFile)
	testza.AssertEqual(t, int64(3), schemaFile.Version)

	schema := GetSchema("SkillGems")
	testza.AssertEqual(t, "SkillGems", schema.Name)
	testza.AssertEqual(t, 17, len(schema.Columns))
}

func TestParseDat(t *testing.T) {
	LoadParser()

	loader, err := GetBundleLoader(NewWebFS())
	testza.AssertNoError(t, err)

	testFiles := []struct {
		Name  string
		Count int
	}{
		{Name: "Data/PassiveTreeExpansionJewels.dat64", Count: 3},
		{Name: "Data/PassiveTreeExpansionSkills.dat64", Count: 55},
		{Name: "Data/PassiveTreeExpansionSpecialSkills.dat64", Count: 307},
		{Name: "Data/CostTypes.dat64", Count: 13},
		{Name: "Data/Mods.dat64", Count: 30516},
		{Name: "Data/ActiveSkills.dat64", Count: 1367},
		{Name: "Data/Essences.dat64", Count: 105},
		{Name: "Data/CraftingBenchOptions.dat64", Count: 835},
		{Name: "Data/PantheonPanelLayout.dat64", Count: 16},
		{Name: "Data/WeaponTypes.dat64", Count: 359},
		{Name: "Data/ArmourTypes.dat64", Count: 421},
		{Name: "Data/ShieldTypes.dat64", Count: 98},
		{Name: "Data/Flasks.dat64", Count: 51},
		{Name: "Data/ComponentCharges.dat64", Count: 51},
		{Name: "Data/ComponentAttributeRequirements.dat64", Count: 795},
		{Name: "Data/BaseItemTypes.dat64", Count: 9479},
		{Name: "Data/Stats.dat64", Count: 16251},
		{Name: "Data/AlternatePassiveSkills.dat64", Count: 155},
		{Name: "Data/AlternatePassiveAdditions.dat64", Count: 94},
		{Name: "Data/DefaultMonsterStats.dat64", Count: 100},
		{Name: "Data/SkillTotemVariations.dat64", Count: 220},
		{Name: "Data/MonsterVarieties.dat64", Count: 8912},
		{Name: "Data/MonsterMapDifficulty.dat64", Count: 25},
		{Name: "Data/MonsterMapBossDifficulty.dat64", Count: 25},
		{Name: "Data/GrantedEffects.dat64", Count: 10020},
		{Name: "Data/SkillTotems.dat64", Count: 20},
		{Name: "Data/GrantedEffectStatSetsPerLevel.dat64", Count: 46717},
		{Name: "Data/GrantedEffectsPerLevel.dat64", Count: 46716},
		{Name: "Data/GrantedEffectQualityStats.dat64", Count: 1459},
		{Name: "Data/SkillGems.dat64", Count: 703},
		{Name: "Data/ItemExperiencePerLevel.dat64", Count: 11432},
		{Name: "Data/Tags.dat64", Count: 1003},
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
