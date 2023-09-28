package poe

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/Vilsol/slox"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/utils"
)

type InitFunction func(ctx context.Context, version string, assetCache loader.AssetCache) error

type initBlock struct {
	Func InitFunction
	Name string
}

func (i initBlock) Load(ctx context.Context, version string, assetCache loader.AssetCache) error {
	slox.From(ctx).Debug("running initialization", slog.String("name", i.Name))
	start := time.Now()
	if err := i.Func(ctx, version, assetCache); err != nil {
		return errors.Wrap(err, "failed to initialize: "+i.Name)
	}
	slox.From(ctx).Debug("running initialization", slog.String("func", i.Name), slog.Duration("took", time.Since(start)))
	return nil
}

var initFunctions = []initBlock{
	{
		Func: InitializeActiveSkillTypes,
		Name: "ActiveSkillTypes",
	},
	{
		Func: InitializeActiveSkills,
		Name: "ActiveSkills",
	},
	{
		Func: InitializeAlternatePassiveAdditions,
		Name: "AlternatePassiveAdditions",
	},
	{
		Func: InitializeAlternatePassiveSkills,
		Name: "AlternatePassiveSkills",
	},
	{
		Func: InitializeArmourTypes,
		Name: "ArmourTypes",
	},
	{
		Func: InitializeBaseItemTypes,
		Name: "BaseItemTypes",
	},
	{
		Func: InitializeComponentAttributeRequirements,
		Name: "ComponentAttributeRequirements",
	},
	{
		Func: InitializeComponentCharges,
		Name: "ComponentCharges",
	},
	{
		Func: InitializeCostTypes,
		Name: "CostTypes",
	},
	{
		Func: InitializeCraftingBenchOptions,
		Name: "CraftingBenchOptions",
	},
	{
		Func: InitializeDefaultMonsterStats,
		Name: "DefaultMonsterStats",
	},
	{
		Func: InitializeEssences,
		Name: "Essences",
	},
	{
		Func: InitializeFlasks,
		Name: "Flasks",
	},
	{
		Func: InitializeGrantedEffectQualityStats,
		Name: "GrantedEffectQualityStats",
	},
	{
		Func: InitializeGrantedEffectStatSets,
		Name: "GrantedEffectStatSets",
	},
	{
		Func: InitializeGrantedEffectStatSetsPerLevels,
		Name: "GrantedEffectStatSetsPerLevels",
	},
	{
		Func: InitializeGrantedEffects,
		Name: "GrantedEffects",
	},
	{
		Func: InitializeGrantedEffectsPerLevels,
		Name: "GrantedEffectsPerLevels",
	},
	{
		Func: InitializeItemClasses,
		Name: "ItemClasses",
	},
	{
		Func: InitializeItemExperiencePerLevels,
		Name: "ItemExperiencePerLevels",
	},
	{
		Func: InitializeMods,
		Name: "Mods",
	},
	{
		Func: InitializeMonsterMapBossDifficulties,
		Name: "MonsterMapBossDifficulties",
	},
	{
		Func: InitializeMonsterMapDifficulties,
		Name: "MonsterMapDifficulties",
	},
	{
		Func: InitializeMonsterVarieties,
		Name: "MonsterVarieties",
	},
	{
		Func: InitializePantheonPanelLayouts,
		Name: "PantheonPanelLayouts",
	},
	{
		Func: InitializePassiveTreeExpansionJewels,
		Name: "PassiveTreeExpansionJewels",
	},
	{
		Func: InitializePassiveTreeExpansionSkills,
		Name: "PassiveTreeExpansionSkills",
	},
	{
		Func: InitializePassiveTreeExpansionSpecialSkills,
		Name: "PassiveTreeExpansionSpecialSkills",
	},
	{
		Func: InitializeShieldTypes,
		Name: "ShieldTypes",
	},
	{
		Func: InitializeSkillGems,
		Name: "SkillGems",
	},
	{
		Func: InitializeSkillTotemVariations,
		Name: "SkillTotemVariations",
	},
	{
		Func: InitializeSkillTotems,
		Name: "SkillTotems",
	},
	{
		Func: InitializeStats,
		Name: "Stats",
	},
	{
		Func: InitializeTags,
		Name: "Tags",
	},
	{
		Func: InitializeWeaponTypes,
		Name: "WeaponTypes",
	},
}

type UpdateFunc func(data string)

var alreadyInitialized = false

func InitializeAll(ctx context.Context, version string, assetCache loader.AssetCache, updateFunc UpdateFunc) error {
	if alreadyInitialized {
		return nil
	}
	alreadyInitialized = true

	if runtime.GOMAXPROCS(0) == 1 {
		for _, function := range initFunctions {
			if updateFunc != nil {
				updateFunc(function.Name)
			}
			if err := function.Load(ctx, version, assetCache); err != nil {
				return err
			}
		}
	} else {
		g := new(errgroup.Group)
		for _, function := range initFunctions {
			fn := function
			g.Go(func() error {
				return fn.Load(ctx, version, assetCache)
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}
	}

	for _, hook := range utils.RegisteredHooks {
		hook()
	}

	return nil
}
