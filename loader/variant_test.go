package loader_test

import (
	"context"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/fields"
	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/go-pob-data/testdata"
)

type noCache struct{}

func (noCache) Get(_ string) ([]byte, error) { return nil, nil }
func (noCache) Set(_ string, _ []byte) error { return nil }
func (noCache) Exists(_ string) bool         { return false }

func TestWithVariantsReadsMin(t *testing.T) {
	cdn := testdata.ServeCommittedData(t)
	version := "3.29"
	minCtx := loader.WithVariants(context.Background(), fields.MinVariant)

	full, err := loader.LoadRaw[*raw.Mod](context.Background(), version, "Mods", nil, noCache{})
	testza.AssertNoError(t, err)
	minRows, err := loader.LoadRaw[*raw.Mod](minCtx, version, "Mods", nil, noCache{})
	testza.AssertNoError(t, err)

	testza.AssertEqual(t, len(full), len(minRows))
	testza.AssertNotEqual(t, "", full[0].ID)
	testza.AssertEqual(t, "", minRows[0].ID, "ID is not listed, so the .min holds zero")
	testza.AssertEqual(t, full[0].Level, minRows[0].Level)

	translation, err := loader.LoadTranslation(minCtx, version, "en", "stat_descriptions", noCache{})
	testza.AssertNoError(t, err)
	testza.AssertNotZero(t, len(translation.Descriptors))

	// Not listed, so read in full.
	_, err = loader.LoadTranslation(minCtx, version, "en", "passive_skill_stat_descriptions", noCache{})
	testza.AssertNoError(t, err)

	testza.AssertEqual(t, []string(nil), cdn.Misses())
}
