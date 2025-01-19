package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
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
	path string
}

func (f PackedFile) FilePath(gameVersion string) string {
	ext := GetFileExtension(gameVersion)
	if IsNewHash(gameVersion) {
		return strings.ToLower(f.path) + "." + ext
	}
	return f.path + "." + ext
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
		path: "Data/PassiveTreeExpansionJewels",
		Fn:   ReEncode[*raw.PassiveTreeExpansionJewel](),
	},
	{
		path: "Data/PassiveTreeExpansionSkills",
		Fn:   ReEncode[*raw.PassiveTreeExpansionSkill](),
	},
	{
		path: "Data/PassiveTreeExpansionSpecialSkills",
		Fn:   ReEncode[*raw.PassiveTreeExpansionSpecialSkill](),
	},
	{
		path: "Data/CostTypes",
		Fn:   ReEncode[*raw.CostType](),
	},
	{
		path: "Data/Mods",
		Fn:   ReEncode[*raw.Mod](),
	},
	{
		path: "Data/ActiveSkills",
		Fn:   ReEncode[*raw.ActiveSkill](),
	},
	{
		path: "Data/Essences",
		Fn:   ReEncode[*raw.Essence](),
	},
	{
		path: "Data/CraftingBenchOptions",
		Fn:   ReEncode[*raw.CraftingBenchOption](),
	},
	{
		path: "Data/PantheonPanelLayout",
		Fn:   ReEncode[*raw.PantheonPanelLayout](),
	},
	{
		path: "Data/WeaponTypes",
		Fn:   ReEncode[*raw.WeaponType](),
	},
	{
		path: "Data/ArmourTypes",
		Fn:   ReEncode[*raw.ArmourType](),
	},
	{
		path: "Data/ShieldTypes",
		Fn:   ReEncode[*raw.ShieldType](),
	},
	{
		path: "Data/Flasks",
		Fn:   ReEncode[*raw.Flask](),
	},
	{
		path: "Data/ComponentCharges",
		Fn:   ReEncode[*raw.ComponentCharge](),
	},
	{
		path: "Data/ComponentAttributeRequirements",
		Fn:   ReEncode[*raw.ComponentAttributeRequirement](),
	},
	{
		path: "Data/BaseItemTypes",
		Fn:   ReEncode[*raw.BaseItemType](),
	},
	{
		path: "Data/Stats",
		Fn:   ReEncode[*raw.Stat](),
	},
	{
		path: "Data/AlternatePassiveSkills",
		Fn:   ReEncode[*raw.AlternatePassiveSkill](),
	},
	{
		path: "Data/AlternatePassiveAdditions",
		Fn:   ReEncode[*raw.AlternatePassiveAddition](),
	},
	{
		path: "Data/DefaultMonsterStats",
		Fn:   ReEncode[*raw.DefaultMonsterStat](),
	},
	{
		path: "Data/SkillTotemVariations",
		Fn:   ReEncode[*raw.SkillTotemVariation](),
	},
	{
		path: "Data/MonsterVarieties",
		Fn:   ReEncode[*raw.MonsterVariety](),
	},
	{
		path: "Data/MonsterMapDifficulty",
		Fn:   ReEncode[*raw.MonsterMapDifficulty](),
	},
	{
		path: "Data/MonsterMapBossDifficulty",
		Fn:   ReEncode[*raw.MonsterMapBossDifficulty](),
	},
	{
		path: "Data/GrantedEffects",
		Fn:   ReEncode[*raw.GrantedEffect](),
	},
	{
		path: "Data/SkillTotems",
		Fn:   ReEncode[*raw.SkillTotem](),
	},
	{
		path: "Data/GrantedEffectStatSetsPerLevel",
		Fn:   ReEncode[*raw.GrantedEffectStatSetsPerLevel](),
	},
	{
		path: "Data/GrantedEffectsPerLevel",
		Fn:   ReEncode[*raw.GrantedEffectsPerLevel](),
	},
	{
		path: "Data/GrantedEffectQualityStats",
		Fn:   ReEncode[*raw.GrantedEffectQualityStat](),
	},
	{
		path: "Data/SkillGems",
		Fn:   ReEncode[*raw.SkillGem](),
	},
	{
		path: "Data/ItemExperiencePerLevel",
		Fn:   ReEncode[*raw.ItemExperiencePerLevel](),
	},
	{
		path: "Data/Tags",
		Fn:   ReEncode[*raw.Tag](),
	},
	{
		path: "Data/ActiveSkillType",
		Fn:   ReEncode[*raw.ActiveSkillType](),
	},
	{
		path: "Data/ItemClasses",
		Fn:   ReEncode[*raw.ItemClass](),
	},
	{
		path: "Data/GrantedEffectStatSets",
		Fn:   ReEncode[*raw.GrantedEffectStatSet](),
	},
	{
		path: "Data/PassiveSkills",
		Fn:   ReEncode[*raw.PassiveSkill](),
	},
	{
		path: "Data/AlternateTreeVersions",
		Fn:   ReEncode[*raw.AlternateTreeVersion](),
	},
	{
		path: "Data/AtlasNode",
		Fn:   ReEncode[*raw.AtlasNode](),
	},
	{
		path: "Data/AtlasNodeDefinition",
		Fn:   ReEncode[*raw.AtlasNodeDefinition](),
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
	loader, err := extractor.GetBundleLoader(os.DirFS(gamePath), IsNewHash(gameVersion))
	if err != nil {
		return errors.Wrap(err, "could not initialize bundle loader")
	}

	for _, file := range filesToExport {
		pathWithExt := file.path + "." + GetFileExtension(gameVersion)

		slog.Info("extracting", slog.String("path", pathWithExt))

		data, err := loader.Open(file.FilePath(gameVersion))
		if err != nil {
			slog.Error("failed to open dat file", slog.String("error", err.Error()), slog.String("path", pathWithExt))
			continue
		}

		dat, err := extractor.ParseDat(data, filepath.Base(pathWithExt))
		if err != nil {
			slog.Error("failed to parse dat file", slog.String("error", err.Error()), slog.String("path", pathWithExt))
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

		outNameGzip := strings.Split(filepath.Base(pathWithExt), ".")[0] + ".json.gz"
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

		outNameBrotli := strings.Split(filepath.Base(pathWithExt), ".")[0] + ".json.br"
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

		outNameMsgpBrotli := strings.Split(filepath.Base(pathWithExt), ".")[0] + ".msgpack.br"
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
		slog.Info("extracting", slog.String("path", img))

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

		slog.Info("compressing", slog.String("path", img))
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
