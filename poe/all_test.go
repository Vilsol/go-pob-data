package poe

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/testdata"
	"github.com/Vilsol/go-pob-data/utils"
)

type FakeCache struct{}

func (f FakeCache) Get(_ string) ([]byte, error) {
	return nil, nil
}

func (f FakeCache) Set(_ string, _ []byte) error {
	return nil
}

func (f FakeCache) Exists(_ string) bool {
	return false
}

func TestAll(t *testing.T) {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))
	testdata.ServeCommittedData(t)

	err := InitializeAll(context.Background(), testdata.LatestDataVersion(), FakeCache{}, nil)
	testza.AssertNoError(t, err)

	effect := GrantedEffectByID("PlayerMelee")
	testza.AssertEqual(t, false, effect.IsSupport)
	testza.AssertEqual(t, 1000, effect.CastTime)
	testza.AssertEqual(t, 4, effect.Attribute)
	testza.AssertEqual(t, utils.Ptr(108), effect.ActiveSkill)

	skill := effect.GetActiveSkill()
	testza.AssertEqual(t, "melee", skill.ID)
	testza.AssertEqual(t, true, skill.IsManuallyCasted)
	testza.AssertEqual(t, "Default Attack", skill.DisplayedName)

	types := skill.GetActiveSkillTypes()
	testza.AssertEqual(t, 8, len(types))
}

func BenchmarkAll(b *testing.B) {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.Level(9)})
	slog.SetDefault(slog.New(h))
	testdata.ServeCommittedData(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alreadyInitialized = false
		_ = InitializeAll(context.Background(), testdata.LatestDataVersion(), nil, nil)
	}
}

// TestInitializeAllRequestsOnlyPresentTables catches default initializers for tables the game no
// longer ships (SkillTotems is gone since 3.26), which cost a 404 on every cold load.
func TestInitializeAllRequestsOnlyPresentTables(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)

	testza.AssertNoError(t, initializeFunctions(context.Background(), initFunctions, testdata.LatestDataVersion(), FakeCache{}, nil))
	testza.AssertEqual(t, []string(nil), cdn.Misses())
}

// go-pob reads the .min assets by loading under loader.WithVariants(ctx, fields.MinVariant).
func TestInitializeAllReadsMinAssets(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	ctx := loader.WithVariants(context.Background(), fields.MinVariant)

	// Later tests read the full tables.
	t.Cleanup(func() {
		testza.AssertNoError(t, initializeFunctions(context.Background(), initFunctions, testdata.LatestDataVersion(), FakeCache{}, nil))
	})

	testza.AssertNoError(t, initializeFunctions(ctx, initFunctions, testdata.LatestDataVersion(), FakeCache{}, nil))
	testza.AssertEqual(t, []string(nil), cdn.Misses())
	testza.AssertEqual(t, "", Mods[0].ID, "Mods.ID is not listed, so the .min holds zero")
}
