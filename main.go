package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
	_ "github.com/andybalholm/brotli"
	"github.com/tinylib/msgp/msgp"
	"github.com/yusukebe/go-pngquant"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/go-pob-data/stat_translations"
)

//go:generate msgp -file ./raw

type PackedFile struct {
	Path string
	Fn   func(jsonData []byte, w *msgp.Writer)
}

func ReEncode[T msgp.Encodable]() func(jsonData []byte, w *msgp.Writer) {
	return func(jsonData []byte, w *msgp.Writer) {
		out := make([]T, 0)
		if err := json.Unmarshal(jsonData, &out); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		for _, obj := range out {
			if err := obj.EncodeMsg(w); err != nil {
				println(err.Error())
				os.Exit(1)
				return
			}
		}
	}
}

var filesToExport = []PackedFile{
	{
		Path: "Data/PassiveTreeExpansionJewels.dat64",
		Fn:   ReEncode[*raw.PassiveTreeExpansionJewel](),
	},
	{
		Path: "Data/PassiveTreeExpansionSkills.dat64",
		Fn:   ReEncode[*raw.PassiveTreeExpansionSkill](),
	},
	{
		Path: "Data/PassiveTreeExpansionSpecialSkills.dat64",
		Fn:   ReEncode[*raw.PassiveTreeExpansionSpecialSkill](),
	},
	{
		Path: "Data/CostTypes.dat64",
		Fn:   ReEncode[*raw.CostType](),
	},
	{
		Path: "Data/Mods.dat64",
		Fn:   ReEncode[*raw.Mod](),
	},
	{
		Path: "Data/ActiveSkills.dat64",
		Fn:   ReEncode[*raw.ActiveSkill](),
	},
	{
		Path: "Data/Essences.dat64",
		Fn:   ReEncode[*raw.Essence](),
	},
	{
		Path: "Data/CraftingBenchOptions.dat64",
		Fn:   ReEncode[*raw.CraftingBenchOption](),
	},
	{
		Path: "Data/PantheonPanelLayout.dat64",
		Fn:   ReEncode[*raw.PantheonPanelLayout](),
	},
	{
		Path: "Data/WeaponTypes.dat64",
		Fn:   ReEncode[*raw.WeaponType](),
	},
	{
		Path: "Data/ArmourTypes.dat64",
		Fn:   ReEncode[*raw.ArmourType](),
	},
	{
		Path: "Data/ShieldTypes.dat64",
		Fn:   ReEncode[*raw.ShieldType](),
	},
	{
		Path: "Data/Flasks.dat64",
		Fn:   ReEncode[*raw.Flask](),
	},
	{
		Path: "Data/ComponentCharges.dat64",
		Fn:   ReEncode[*raw.ComponentCharge](),
	},
	{
		Path: "Data/ComponentAttributeRequirements.dat64",
		Fn:   ReEncode[*raw.ComponentAttributeRequirement](),
	},
	{
		Path: "Data/BaseItemTypes.dat64",
		Fn:   ReEncode[*raw.BaseItemType](),
	},
	{
		Path: "Data/Stats.dat64",
		Fn:   ReEncode[*raw.Stat](),
	},
	{
		Path: "Data/AlternatePassiveSkills.dat64",
		Fn:   ReEncode[*raw.AlternatePassiveSkill](),
	},
	{
		Path: "Data/AlternatePassiveAdditions.dat64",
		Fn:   ReEncode[*raw.AlternatePassiveAddition](),
	},
	{
		Path: "Data/DefaultMonsterStats.dat64",
		Fn:   ReEncode[*raw.DefaultMonsterStat](),
	},
	{
		Path: "Data/SkillTotemVariations.dat64",
		Fn:   ReEncode[*raw.SkillTotemVariation](),
	},
	{
		Path: "Data/MonsterVarieties.dat64",
		Fn:   ReEncode[*raw.MonsterVariety](),
	},
	{
		Path: "Data/MonsterMapDifficulty.dat64",
		Fn:   ReEncode[*raw.MonsterMapDifficulty](),
	},
	{
		Path: "Data/MonsterMapBossDifficulty.dat64",
		Fn:   ReEncode[*raw.MonsterMapBossDifficulty](),
	},
	{
		Path: "Data/GrantedEffects.dat64",
		Fn:   ReEncode[*raw.GrantedEffect](),
	},
	{
		Path: "Data/SkillTotems.dat64",
		Fn:   ReEncode[*raw.SkillTotem](),
	},
	{
		Path: "Data/GrantedEffectStatSetsPerLevel.dat64",
		Fn:   ReEncode[*raw.GrantedEffectStatSetsPerLevel](),
	},
	{
		Path: "Data/GrantedEffectsPerLevel.dat64",
		Fn:   ReEncode[*raw.GrantedEffectsPerLevel](),
	},
	{
		Path: "Data/GrantedEffectQualityStats.dat64",
		Fn:   ReEncode[*raw.GrantedEffectQualityStat](),
	},
	{
		Path: "Data/SkillGems.dat64",
		Fn:   ReEncode[*raw.SkillGem](),
	},
	{
		Path: "Data/ItemExperiencePerLevel.dat64",
		Fn:   ReEncode[*raw.ItemExperiencePerLevel](),
	},
	{
		Path: "Data/Tags.dat64",
		Fn:   ReEncode[*raw.Tag](),
	},
	{
		Path: "Data/ActiveSkillType.dat64",
		Fn:   ReEncode[*raw.ActiveSkillType](),
	},
	{
		Path: "Data/ItemClasses.dat64",
		Fn:   ReEncode[*raw.ItemClass](),
	},
	{
		Path: "Data/GrantedEffectStatSets.dat64",
		Fn:   ReEncode[*raw.GrantedEffectStatSet](),
	},
	{
		Path: "Data/PassiveSkills.dat64",
		Fn:   ReEncode[*raw.PassiveSkill](),
	},
	{
		Path: "Data/AlternateTreeVersions.dat64",
		Fn:   ReEncode[*raw.AlternateTreeVersion](),
	},
}

const GGGRepoBase = "https://raw.githubusercontent.com/grindinggear/skilltree-export/%s/"

var skillTreeSpriteGroups = []string{
	"background",
	"normalActive",
	"notableActive",
	"keystoneActive",
	"normalInactive",
	"notableInactive",
	"keystoneInactive",
	"mastery",
	"masteryConnected",
	"masteryActiveSelected",
	"masteryInactive",
	"masteryActiveEffect",
	"ascendancyBackground",
	"ascendancy",
	"startNode",
	"groupBackground",
	"frame",
	"jewel",
	"line",
	"jewelRadius",
}

var translations = []string{
	"Metadata/StatDescriptions/stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_aura_stat_descriptions.txt",
	"Metadata/StatDescriptions/passive_skill_stat_descriptions.txt",
}

var imagesToExport = []string{
	"Art/2DArt/BaseClassIllustrations/Str.dds",
	"Art/2DArt/BaseClassIllustrations/Dex.dds",
	"Art/2DArt/BaseClassIllustrations/Int.dds",
	"Art/2DArt/BaseClassIllustrations/StrDex.dds",
	"Art/2DArt/BaseClassIllustrations/StrInt.dds",
	"Art/2DArt/BaseClassIllustrations/DexInt.dds",
}

func main() {
	if len(os.Args) < 2 {
		println("please provide path to the game directory")
		os.Exit(1)
		return
	}

	if len(os.Args) < 3 {
		println("please provide passive tree version")
		os.Exit(1)
		return
	}

	if len(os.Args) < 4 {
		println("please provide game version")
		os.Exit(1)
		return
	}

	gamePath := os.Args[1]
	treeVersion := os.Args[2]
	gameVersion := os.Args[3]

	extractRawData(gamePath, gameVersion)
	downloadTreeData(treeVersion, gameVersion)
	extractTranslations(gamePath, gameVersion)
}

func extractRawData(gamePath string, gameVersion string) {
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
		println("Extracting", file.Path)

		data, err := loader.Open(file.Path)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		dat, err := extractor.ParseDat(data, filepath.Base(file.Path))
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		// Ensure that all slices are always sorted by key
		sort.Slice(dat, func(i, j int) bool {
			return reflect.ValueOf(dat[i]).Field(0).Int() < reflect.ValueOf(dat[j]).Field(0).Int()
		})

		b, err := json.Marshal(dat)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		outNameGzip := strings.Split(filepath.Base(file.Path), ".")[0] + ".json.gz"
		outPathGzip := filepath.Join("data", gameVersion, "raw", outNameGzip)

		if err := os.MkdirAll(filepath.Dir(outPathGzip), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		fGzip, err := os.OpenFile(outPathGzip, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writerGzip := gzip.NewWriter(fGzip)
		if _, err := writerGzip.Write(b); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writerGzip.Close()
		_ = fGzip.Close()

		outNameBrotli := strings.Split(filepath.Base(file.Path), ".")[0] + ".json.br"
		outPathBrotli := filepath.Join("data", gameVersion, "raw", outNameBrotli)

		fBrotli, err := os.OpenFile(outPathBrotli, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writerBrotli := brotli.NewWriter(fBrotli)
		if _, err := writerBrotli.Write(b); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writerBrotli.Close()
		_ = fBrotli.Close()

		outNameMsgpBrotli := strings.Split(filepath.Base(file.Path), ".")[0] + ".msgpack.br"
		outPathMsgpBrotli := filepath.Join("data", gameVersion, "raw", outNameMsgpBrotli)

		fMsgpBrotli, err := os.OpenFile(outPathMsgpBrotli, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		writerMsgpBrotli := brotli.NewWriter(fMsgpBrotli)

		msgpWriter := msgp.NewWriter(writerMsgpBrotli)
		file.Fn(b, msgpWriter)

		if err := msgpWriter.Flush(); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		_ = writerMsgpBrotli.Close()
		_ = fMsgpBrotli.Close()

	}

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()

	for _, img := range imagesToExport {
		println("Extracting", img)

		data, err := loader.Open(img)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		fullImage, err := io.ReadAll(data)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		if err := mw.ReadImageBlob(fullImage); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		outName := strings.Split(filepath.Base(img), ".")[0] + ".png"
		outPath := filepath.Join("data", gameVersion, "raw", filepath.Dir(img), outName)

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		if err := mw.SetImageFormat("png"); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		uncompressed := mw.GetImageBlob()

		finalImage, err := pngquant.CompressBytes(uncompressed, "one")
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		if err := os.WriteFile(outPath, finalImage, 0755); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}
	}
}

func downloadTreeData(treeVersion string, gameVersion string) {
	repoVersionBase := fmt.Sprintf(GGGRepoBase, treeVersion)
	response, err := http.DefaultClient.Get(repoVersionBase + "/data.json")
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	defer response.Body.Close()
	dataJSON, err := io.ReadAll(response.Body)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	dataJSONOutPathGzip := filepath.Join("data", gameVersion, "tree", "data.json.gz")

	if err := os.MkdirAll(filepath.Dir(dataJSONOutPathGzip), 0755); err != nil {
		if !os.IsExist(err) {
			println(err.Error())
			os.Exit(1)
			return
		}
	}

	fGzip, err := os.OpenFile(dataJSONOutPathGzip, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	writerGzip := gzip.NewWriter(fGzip)
	if _, err := writerGzip.Write(dataJSON); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	_ = writerGzip.Close()
	_ = fGzip.Close()

	dataJSONOutPathBrotli := filepath.Join("data", gameVersion, "tree", "data.json.br")

	fBrotli, err := os.OpenFile(dataJSONOutPathBrotli, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	writerBrotli := brotli.NewWriter(fBrotli)
	if _, err := writerBrotli.Write(dataJSON); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	_ = writerBrotli.Close()
	_ = fBrotli.Close()

	var skillTreeData SkillTreeData
	if err := json.Unmarshal(dataJSON, &skillTreeData); err != nil {
		println(err.Error())
		os.Exit(1)
		return
	}

	downloaded := make(map[string]bool)

	selectedResolution := skillTreeData.ImageZoomLevels[len(skillTreeData.ImageZoomLevels)-1]
	for _, group := range skillTreeSpriteGroups {
		spriteGroup := skillTreeData.Sprites[group]
		assetPath := spriteGroup[strconv.FormatFloat(selectedResolution, 'f', -1, 64)]
		if assetPath.Filename == "" {
			for _, p := range spriteGroup {
				assetPath = p
				break
			}
		}

		parsedURL, err := url.Parse(assetPath.Filename)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		fileName := path.Base(parsedURL.Path)
		if _, ok := downloaded[fileName]; ok {
			continue
		}
		downloaded[fileName] = true

		println("Downloading", fileName)

		response, err := http.DefaultClient.Get(repoVersionBase + "/assets/" + fileName)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		defer response.Body.Close()
		fileData, err := io.ReadAll(response.Body)
		if err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}

		fileOutPath := filepath.Join("data", gameVersion, "tree", "assets", fileName)
		if err := os.MkdirAll(filepath.Dir(fileOutPath), 0755); err != nil {
			if !os.IsExist(err) {
				println(err.Error())
				os.Exit(1)
				return
			}
		}

		if err := os.WriteFile(fileOutPath, fileData, 0755); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}
	}
}

func extractTranslations(gamePath string, gameVersion string) {
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

	parser := stat_translations.NewTranslationParser(loader)

	for _, translation := range translations {
		if err := parser.ParseFile(translation); err != nil {
			println(err.Error())
			os.Exit(1)
			return
		}
	}
}
