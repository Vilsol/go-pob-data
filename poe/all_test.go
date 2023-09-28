package poe

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"

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

	err := InitializeAll(context.Background(), testdata.ShortGameVersion(), FakeCache{}, nil)
	testza.AssertNoError(t, err)

	effect := GrantedEffectByID("PlayerMelee")
	testza.AssertEqual(t, false, effect.IsSupport)
	testza.AssertEqual(t, 1000, effect.CastTime)
	testza.AssertEqual(t, 4, effect.Attribute)
	testza.AssertEqual(t, utils.Ptr(65), effect.ActiveSkill)

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		alreadyInitialized = false
		_ = InitializeAll(context.Background(), testdata.ShortGameVersion(), nil, nil)
	}
}
