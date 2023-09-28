package loader

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

func TestTranslationLoader(t *testing.T) {
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(h))

	_, err := LoadTranslation(context.Background(), testdata.ShortGameVersion(), "en", "stat_descriptions", FakeCache{})
	testza.AssertNoError(t, err)
}
