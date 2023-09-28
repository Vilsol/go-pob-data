package loader

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/go-pob-data/testdata"
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

func TestRawLoader(t *testing.T) {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	_, err := LoadRaw[*raw.GrantedEffectStatSet](context.Background(), testdata.ShortGameVersion(), "GrantedEffectStatSets", nil, FakeCache{})
	testza.AssertNoError(t, err)
}
