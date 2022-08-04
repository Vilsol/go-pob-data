package main

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Vilsol/go-pob-data/extractor"
)

var filesToExport = []string{
	"Data/PassiveTreeExpansionJewels.dat64",
	"Data/PassiveTreeExpansionSkills.dat64",
	"Data/PassiveTreeExpansionSpecialSkills.dat64",
	"Data/CostTypes.dat64",
	"Data/Mods.dat64",
	"Data/ActiveSkills.dat64",
	"Data/Essences.dat64",
	"Data/CraftingBenchOptions.dat64",
	"Data/PantheonPanelLayout.dat64",
	"Data/WeaponTypes.dat64",
	"Data/ArmourTypes.dat64",
	"Data/ShieldTypes.dat64",
	"Data/Flasks.dat64",
	"Data/ComponentCharges.dat64",
	"Data/ComponentAttributeRequirements.dat64",
	"Data/BaseItemTypes.dat64",
	"Data/Stats.dat64",
	"Data/AlternatePassiveSkills.dat64",
	"Data/AlternatePassiveAdditions.dat64",
	"Data/DefaultMonsterStats.dat64",
	"Data/SkillTotemVariations.dat64",
	"Data/MonsterVarieties.dat64",
	"Data/MonsterMapDifficulty.dat64",
	"Data/MonsterMapBossDifficulty.dat64",
	"Data/GrantedEffects.dat64",
	"Data/SkillTotems.dat64",
	"Data/GrantedEffectStatSetsPerLevel.dat64",
	"Data/GrantedEffectsPerLevel.dat64",
	"Data/GrantedEffectQualityStats.dat64",
	"Data/SkillGems.dat64",
	"Data/ItemExperiencePerLevel.dat64",
	"Data/Tags.dat64",
	"Data/ActiveSkillType.dat64",
	"Data/ItemClasses.dat64",
	"Data/GrantedEffectStatSets.dat64",
}

func main() {
	if len(os.Args) < 2 {
		println("please provide path to the game directory")
		os.Exit(1)
		return
	}

	if len(os.Args) < 3 {
		println("please provide game version")
		os.Exit(1)
		return
	}

	gamePath := os.Args[1]
	gameVersion := os.Args[2]

	if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	extractor.LoadParser()
	loader, err := extractor.GetBundleLoader(os.DirFS(gamePath))
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	for _, file := range filesToExport {
		println("Extracting", file)

		data, err := loader.Open(file)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		dat, err := extractor.ParseDat(data, filepath.Base(file))
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		b, err := json.Marshal(dat)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		outName := strings.Split(filepath.Base(file), ".")[0] + ".json.gz"
		outPath := filepath.Join("data", gameVersion, outName)

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writer := gzip.NewWriter(f)
		if _, err := writer.Write(b); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writer.Close()
		_ = f.Close()
	}
}
