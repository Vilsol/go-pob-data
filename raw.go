package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
	"gopkg.in/gographics/imagick.v3/imagick"

	"github.com/Vilsol/go-pob-data/extractor"
	"github.com/Vilsol/go-pob-data/raw"
)

type PackedFile struct {
	Fn   func(jsonData []byte, w *msgp.Writer) error
	Path string
}

func ReEncode[T msgp.Encodable]() func(jsonData []byte, w *msgp.Writer) error {
	return func(jsonData []byte, w *msgp.Writer) error {
		out := make([]T, 0)
		if err := json.Unmarshal(jsonData, &out); err != nil {
			return errors.Wrap(err, "failed to unmarshal data")
		}

		for _, obj := range out {
			if err := obj.EncodeMsg(w); err != nil {
				return errors.Wrap(err, "failed to encode msgpack")
			}
		}

		return nil
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

var imagesToExport = []string{
	"Art/2DArt/BaseClassIllustrations/Str.dds",
	"Art/2DArt/BaseClassIllustrations/Dex.dds",
	"Art/2DArt/BaseClassIllustrations/Int.dds",
	"Art/2DArt/BaseClassIllustrations/StrDex.dds",
	"Art/2DArt/BaseClassIllustrations/StrInt.dds",
	"Art/2DArt/BaseClassIllustrations/DexInt.dds",
}

func extractRawData(gamePath string, gameVersion string) error {
	if _, err := os.Stat(filepath.Join(gamePath, "Bundles2", "_.index.bin")); err != nil {
		return errors.Wrap(err, "could not find _.index.bin")
	}

	extractor.LoadParser(gameVersion)
	loader, err := extractor.GetBundleLoader(os.DirFS(gamePath))
	if err != nil {
		return errors.Wrap(err, "could not initialize bundle loader")
	}

	for _, file := range filesToExport {
		println("Extracting", file.Path)

		data, err := loader.Open(file.Path)
		if err != nil {
			return errors.Wrap(err, "could not open file")
		}

		dat, err := extractor.ParseDat(data, filepath.Base(file.Path))
		if err != nil {
			fmt.Println(errors.Wrap(err, "failed to parse dat file"))
			continue
		}

		// Ensure that all slices are always sorted by key
		sort.Slice(dat, func(i, j int) bool {
			return reflect.ValueOf(dat[i]).Field(0).Int() < reflect.ValueOf(dat[j]).Field(0).Int()
		})

		b, err := json.Marshal(dat)
		if err != nil {
			return errors.Wrap(err, "failed to marshal dat file")
		}

		outNameGzip := strings.Split(filepath.Base(file.Path), ".")[0] + ".json.gz"
		outPathGzip := filepath.Join("data", gameVersion, "raw", outNameGzip)

		if err := os.MkdirAll(filepath.Dir(outPathGzip), 0o755); err != nil {
			if !os.IsExist(err) {
				return errors.Wrap(err, "could not create directory tree")
			}
		}

		fGzip, err := os.OpenFile(outPathGzip, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return errors.Wrap(err, "could not open file for writing")
		}

		writerGzip := gzip.NewWriter(fGzip)
		if _, err := writerGzip.Write(b); err != nil {
			return errors.Wrap(err, "could not compress to gzip")
		}

		_ = writerGzip.Close()
		_ = fGzip.Close()

		outNameBrotli := strings.Split(filepath.Base(file.Path), ".")[0] + ".json.br"
		outPathBrotli := filepath.Join("data", gameVersion, "raw", outNameBrotli)

		fBrotli, err := os.OpenFile(outPathBrotli, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return errors.Wrap(err, "could not open file for writing")
		}

		writerBrotli := brotli.NewWriter(fBrotli)
		if _, err := writerBrotli.Write(b); err != nil {
			return errors.Wrap(err, "could not compress to brotli")
		}

		_ = writerBrotli.Close()
		_ = fBrotli.Close()

		outNameMsgpBrotli := strings.Split(filepath.Base(file.Path), ".")[0] + ".msgpack.br"
		outPathMsgpBrotli := filepath.Join("data", gameVersion, "raw", outNameMsgpBrotli)

		fMsgpBrotli, err := os.OpenFile(outPathMsgpBrotli, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
		if err != nil {
			return errors.Wrap(err, "could not open file for writing")
		}

		writerMsgpBrotli := brotli.NewWriter(fMsgpBrotli)

		msgpWriter := msgp.NewWriter(writerMsgpBrotli)
		_ = msgpWriter.WriteInt64(int64(len(dat)))
		if err := file.Fn(b, msgpWriter); err != nil {
			return errors.Wrap(err, "could not write to msgpack")
		}

		if err := msgpWriter.Flush(); err != nil {
			return errors.Wrap(err, "could not flush msgpack")
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
			return errors.Wrap(err, "could not open file")
		}

		fullImage, err := io.ReadAll(data)
		if err != nil {
			return errors.Wrap(err, "failed to read file contents")
		}

		if err := mw.ReadImageBlob(fullImage); err != nil {
			return errors.Wrap(err, "failed to read image as a blob")
		}

		outName := strings.Split(filepath.Base(img), ".")[0] + ".png"
		outPath := filepath.Join("data", gameVersion, "raw", filepath.Dir(img), outName)

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			if !os.IsExist(err) {
				return errors.Wrap(err, "failed to create directory tree")
			}
		}

		if err := mw.SetImageFormat("png"); err != nil {
			return errors.Wrap(err, "failed to change image format")
		}

		uncompressed := mw.GetImageBlob()

		finalImage, err := CompressBytes(uncompressed, "1")
		if err != nil {
			return errors.Wrap(err, "failed to crush png image")
		}

		if err := os.WriteFile(outPath, finalImage, 0o755); err != nil {
			return errors.Wrap(err, "failed to write image to file")
		}
	}
	return nil
}

func CompressBytes(input []byte, speed string) ([]byte, error) {
	cmd := exec.Command("pngquant", "-", "--speed", speed)
	cmd.Stdin = strings.NewReader(string(input))

	var o bytes.Buffer
	cmd.Stdout = &o

	var e bytes.Buffer
	cmd.Stderr = &e

	err := cmd.Run()

	return o.Bytes(), errors.Wrap(err, e.String())
}
