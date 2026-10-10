package poe

import (
	"context"
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/testdata"
)

// initGameData loads the default tables and fns, the initializers of tables that are not part of
// InitializeAll: consumers load those only when they derive PoB's data files from them.
func initGameData(t *testing.T, fns ...InitFunction) {
	t.Helper()
	cdn := testdata.ServeCommittedData(t)
	ctx := context.Background()
	version := testdata.LatestDataVersion()

	testza.AssertNoError(t, initializeFunctions(ctx, initFunctions, version, FakeCache{}, nil))
	for _, fn := range fns {
		testza.AssertNoError(t, fn(ctx, version, FakeCache{}))
	}
	testza.AssertEqual(t, []string(nil), cdn.Misses())
}
