package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
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
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PackedFile struct {
	Codec
	path string
}

type Codec struct {
	// Encode turns a table's json into its brotli tuple asset: an int64 row count, then the rows.
	Encode  func(jsonData []byte) ([]byte, error)
	Variant string
}

func (f PackedFile) FilePath(gameVersion string) string {
	ext := GetFileExtension(gameVersion)
	if IsNewHash(gameVersion) {
		return strings.ToLower(f.path) + "." + ext
	}
	return f.path + "." + ext
}

func ReEncode[T msgp.Encodable]() Codec {
	var row T
	return Codec{
		Variant: loader.TupleVariant(row),
		Encode: func(jsonData []byte) ([]byte, error) {
			out := make([]T, 0)
			if err := json.Unmarshal(jsonData, &out); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal data")
			}

			var buf bytes.Buffer
			bw := brotli.NewWriter(&buf)
			w := msgp.NewWriter(bw)
			if err := w.WriteInt64(int64(len(out))); err != nil {
				return nil, errors.Wrap(err, "failed to encode msgpack")
			}

			for _, obj := range out {
				if err := obj.EncodeMsg(w); err != nil {
					return nil, errors.Wrap(err, "failed to encode msgpack")
				}
			}

			if err := w.Flush(); err != nil {
				return nil, errors.Wrap(err, "failed to flush msgpack")
			}

			if err := bw.Close(); err != nil {
				return nil, errors.Wrap(err, "failed to compress msgpack")
			}

			return buf.Bytes(), nil
		},
	}
}

var filesToExport = []PackedFile{
	{
		path:  "Data/PassiveTreeExpansionJewels",
		Codec: ReEncode[*raw.PassiveTreeExpansionJewel](),
	},
	{
		path:  "Data/PassiveTreeExpansionSkills",
		Codec: ReEncode[*raw.PassiveTreeExpansionSkill](),
	},
	{
		path:  "Data/PassiveTreeExpansionSpecialSkills",
		Codec: ReEncode[*raw.PassiveTreeExpansionSpecialSkill](),
	},
	{
		path:  "Data/CostTypes",
		Codec: ReEncode[*raw.CostType](),
	},
	{
		path:  "Data/Mods",
		Codec: ReEncode[*raw.Mod](),
	},
	{
		path:  "Data/ActiveSkills",
		Codec: ReEncode[*raw.ActiveSkill](),
	},
	{
		path:  "Data/Essences",
		Codec: ReEncode[*raw.Essence](),
	},
	{
		path:  "Data/CraftingBenchOptions",
		Codec: ReEncode[*raw.CraftingBenchOption](),
	},
	{
		path:  "Data/PantheonPanelLayout",
		Codec: ReEncode[*raw.PantheonPanelLayout](),
	},
	{
		path:  "Data/WeaponTypes",
		Codec: ReEncode[*raw.WeaponType](),
	},
	{
		path:  "Data/ArmourTypes",
		Codec: ReEncode[*raw.ArmourType](),
	},
	{
		path:  "Data/ShieldTypes",
		Codec: ReEncode[*raw.ShieldType](),
	},
	{
		path:  "Data/Flasks",
		Codec: ReEncode[*raw.Flask](),
	},
	{
		path:  "Data/ComponentCharges",
		Codec: ReEncode[*raw.ComponentCharge](),
	},
	{
		path:  "Data/ComponentAttributeRequirements",
		Codec: ReEncode[*raw.ComponentAttributeRequirement](),
	},
	{
		path:  "Data/BaseItemTypes",
		Codec: ReEncode[*raw.BaseItemType](),
	},
	{
		path:  "Data/Stats",
		Codec: ReEncode[*raw.Stat](),
	},
	{
		path:  "Data/AlternatePassiveSkills",
		Codec: ReEncode[*raw.AlternatePassiveSkill](),
	},
	{
		path:  "Data/AlternatePassiveAdditions",
		Codec: ReEncode[*raw.AlternatePassiveAddition](),
	},
	{
		path:  "Data/DefaultMonsterStats",
		Codec: ReEncode[*raw.DefaultMonsterStat](),
	},
	{
		path:  "Data/SkillTotemVariations",
		Codec: ReEncode[*raw.SkillTotemVariation](),
	},
	{
		path:  "Data/MonsterVarieties",
		Codec: ReEncode[*raw.MonsterVariety](),
	},
	{
		path:  "Data/MonsterMapDifficulty",
		Codec: ReEncode[*raw.MonsterMapDifficulty](),
	},
	{
		path:  "Data/MonsterMapBossDifficulty",
		Codec: ReEncode[*raw.MonsterMapBossDifficulty](),
	},
	{
		path:  "Data/GrantedEffects",
		Codec: ReEncode[*raw.GrantedEffect](),
	},
	{
		path:  "Data/SkillTotems",
		Codec: ReEncode[*raw.SkillTotem](),
	},
	{
		path:  "Data/GrantedEffectStatSetsPerLevel",
		Codec: ReEncode[*raw.GrantedEffectStatSetsPerLevel](),
	},
	{
		path:  "Data/GrantedEffectsPerLevel",
		Codec: ReEncode[*raw.GrantedEffectsPerLevel](),
	},
	{
		path:  "Data/GrantedEffectQualityStats",
		Codec: ReEncode[*raw.GrantedEffectQualityStat](),
	},
	{
		path:  "Data/SkillGems",
		Codec: ReEncode[*raw.SkillGem](),
	},
	{
		path:  "Data/ItemExperiencePerLevel",
		Codec: ReEncode[*raw.ItemExperiencePerLevel](),
	},
	{
		path:  "Data/Tags",
		Codec: ReEncode[*raw.Tag](),
	},
	{
		path:  "Data/ActiveSkillType",
		Codec: ReEncode[*raw.ActiveSkillType](),
	},
	{
		path:  "Data/ItemClasses",
		Codec: ReEncode[*raw.ItemClass](),
	},
	{
		path:  "Data/GrantedEffectStatSets",
		Codec: ReEncode[*raw.GrantedEffectStatSet](),
	},
	{
		path:  "Data/PassiveSkills",
		Codec: ReEncode[*raw.PassiveSkill](),
	},
	{
		path:  "Data/AlternateTreeVersions",
		Codec: ReEncode[*raw.AlternateTreeVersion](),
	},
	{
		path:  "Data/AtlasNode",
		Codec: ReEncode[*raw.AtlasNode](),
	},
	{
		path:  "Data/AtlasNodeDefinition",
		Codec: ReEncode[*raw.AtlasNodeDefinition](),
	},
	{
		path:  "Data/PassiveSkillOverrides",
		Codec: ReEncode[*raw.PassiveSkillOverride](),
	},
	{
		path:  "Data/PassiveSkillOverrideTypes",
		Codec: ReEncode[*raw.PassiveSkillOverrideType](),
	},
	{
		path:  "Data/PassiveSkillTattoos",
		Codec: ReEncode[*raw.PassiveSkillTattoo](),
	},
	{
		path:  "Data/PassiveSkillTattooTargetSets",
		Codec: ReEncode[*raw.PassiveSkillTattooTargetSet](),
	},
	{
		path:  "Data/PassiveOverrideLimits",
		Codec: ReEncode[*raw.PassiveOverrideLimit](),
	},
	{
		path:  "Data/ClientStrings",
		Codec: ReEncode[*raw.ClientString](),
	},
	{
		path:  "Data/CurrencyExchange",
		Codec: ReEncode[*raw.CurrencyExchange](),
	},
	{
		path:  "Data/FlavourText",
		Codec: ReEncode[*raw.FlavourText](),
	},
	{
		path:  "Data/UniqueStashLayout",
		Codec: ReEncode[*raw.UniqueStashLayout](),
	},
	{
		path:  "Data/UniqueStashTypes",
		Codec: ReEncode[*raw.UniqueStashType](),
	},
	{
		path:  "Data/Words",
		Codec: ReEncode[*raw.Word](),
	},
	{
		path:  "Data/ItemVisualIdentity",
		Codec: ReEncode[*raw.ItemVisualIdentity](),
	},
	{
		path:  "Data/WeaponPassiveSkills",
		Codec: ReEncode[*raw.WeaponPassiveSkill](),
	},
	{
		path:  "Data/WeaponPassiveSkillTypes",
		Codec: ReEncode[*raw.WeaponPassiveSkillType](),
	},
	{
		path:  "Data/PassiveJewelSlots",
		Codec: ReEncode[*raw.PassiveJewelSlot](),
	},
	{
		path:  "Data/PassiveTreeExpansionJewelSizes",
		Codec: ReEncode[*raw.PassiveTreeExpansionJewelSize](),
	},
	{
		path:  "Data/MonsterTypes",
		Codec: ReEncode[*raw.MonsterType](),
	},
	{
		path:  "Data/GameConstants",
		Codec: ReEncode[*raw.GameConstant](),
	},
	{
		path:  "Data/VillageBalancePerLevelShared",
		Codec: ReEncode[*raw.VillageBalancePerLevelShared](),
	},
	{
		path:  "Data/AlternateTreeArt",
		Codec: ReEncode[*raw.AlternateTreeArt](),
	},
	{
		path:  "Data/PassiveSkillTreeNodeFrameArt",
		Codec: ReEncode[*raw.PassiveSkillTreeNodeFrameArt](),
	},
	{
		path:  "Data/PassiveSkillTreeConnectionArt",
		Codec: ReEncode[*raw.PassiveSkillTreeConnectionArt](),
	},
	{
		path:  "Data/GemEffects",
		Codec: ReEncode[*raw.GemEffect](),
	},
	{
		path:  "Data/GemTags",
		Codec: ReEncode[*raw.GemTag](),
	},
	{
		path:  "Data/Tinctures",
		Codec: ReEncode[*raw.Tincture](),
	},
	{
		path:  "Data/ModType",
		Codec: ReEncode[*raw.ModType](),
	},
	{
		path:  "Data/ModFamily",
		Codec: ReEncode[*raw.ModFamily](),
	},
	{
		path:  "Data/CraftingItemClassCategories",
		Codec: ReEncode[*raw.CraftingItemClassCategory](),
	},
	{
		path:  "Data/MonsterResistances",
		Codec: ReEncode[*raw.MonsterResistance](),
	},
	{
		path:  "Data/EffectivenessCostConstants",
		Codec: ReEncode[*raw.EffectivenessCostConstant](),
	},
	{
		path:  "Data/BuffTemplates",
		Codec: ReEncode[*raw.BuffTemplate](),
	},
	{
		path:  "Data/VirtualStatContextFlags",
		Codec: ReEncode[*raw.VirtualStatContextFlag](),
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

func extractRawData(bundles fs.FS, gameVersion string) error {
	for _, file := range filesToExport {
		pathWithExt := file.path + "." + GetFileExtension(gameVersion)

		slog.Info("extracting", slog.String("path", pathWithExt))

		data, err := bundles.Open(file.FilePath(gameVersion))
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

		name := strings.Split(filepath.Base(pathWithExt), ".")[0]
		asset, err := file.Encode(b)
		if err != nil {
			return errors.Wrap(err, "could not encode "+name)
		}

		outPathTuple := filepath.Join("data", gameVersion, loader.AssetPath("raw/"+name, file.Variant))
		if err := os.WriteFile(outPathTuple, asset, 0o755); err != nil {
			return errors.Wrap(err, "could not write "+outPathTuple)
		}
	}

	imagick.Initialize()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	for _, img := range imagesToExport {
		slog.Info("extracting", slog.String("path", img))

		data, err := bundles.Open(strings.ToLower(img))
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
