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
	Func    InitFunction
	Name    string
	Depends []string
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
		Func:    InitializeGrantedEffects,
		Name:    "GrantedEffects",
		Depends: []string{"ActiveSkills"},
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

	if err := initializeFunctions(ctx, initFunctions, version, assetCache, updateFunc); err != nil {
		return err
	}

	for _, hook := range utils.RegisteredHooks {
		hook()
	}

	return nil
}

func initializeFunctions(ctx context.Context, functions []initBlock, version string, assetCache loader.AssetCache, updateFunc UpdateFunc) error {
	if runtime.GOMAXPROCS(0) == 1 {
		for _, function := range functions {
			if updateFunc != nil {
				updateFunc(function.Name)
			}
			if err := function.Load(ctx, version, assetCache); err != nil {
				return err
			}
		}

		return nil
	}

	loaded := make(map[string]bool, len(functions))
	for len(loaded) < len(functions) {
		before := len(loaded)

		g := new(errgroup.Group)
		round := make([]string, 0, len(functions))
		for _, function := range functions {
			// Do not load again
			if _, ok := loaded[function.Name]; ok {
				continue
			}

			// Check if all dependencies are loaded
			canLoad := true
			for _, depend := range function.Depends {
				if _, ok := loaded[depend]; !ok {
					canLoad = false
					break
				}
			}

			if !canLoad {
				continue
			}

			round = append(round, function.Name)

			fn := function
			g.Go(func() error {
				return fn.Load(ctx, version, assetCache)
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}

		// Only the round owner touches `loaded`, after every goroutine of the round is done
		for _, name := range round {
			loaded[name] = true
		}

		if len(loaded) == before {
			return errors.New("failed resolving initialization dependency tree")
		}
	}

	return nil
}
