package poe

import (
	"context"
	"runtime"
	"sync"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/loader"
)

func noopBlock(name string, depends ...string) initBlock {
	return initBlock{
		Name:    name,
		Depends: depends,
		Func: func(_ context.Context, _ string, _ loader.AssetCache) error {
			return nil
		},
	}
}

// TestInitializeFunctionsConcurrentRound forces every block of a round to be in flight at the
// same time, which is what made the round bookkeeping race under -race.
func TestInitializeFunctionsConcurrentRound(t *testing.T) {
	if runtime.GOMAXPROCS(0) == 1 {
		t.Skip("requires the concurrent resolution path")
	}

	const count = 8

	barrier := sync.WaitGroup{}
	barrier.Add(count)

	functions := make([]initBlock, 0, count)
	for i := range count {
		functions = append(functions, initBlock{
			Name: string(rune('A' + i)),
			Func: func(_ context.Context, _ string, _ loader.AssetCache) error {
				barrier.Done()
				barrier.Wait()
				return nil
			},
		})
	}

	testza.AssertNoError(t, initializeFunctions(context.Background(), functions, "", nil, nil))
}

func TestInitializeFunctionsDependencyOrder(t *testing.T) {
	lock := sync.Mutex{}
	order := make([]string, 0, 4)

	block := func(name string, depends ...string) initBlock {
		return initBlock{
			Name:    name,
			Depends: depends,
			Func: func(_ context.Context, _ string, _ loader.AssetCache) error {
				lock.Lock()
				defer lock.Unlock()
				order = append(order, name)
				return nil
			},
		}
	}

	// Declared back to front, so only the dependency resolution can produce the right order
	functions := []initBlock{
		block("D", "C"),
		block("C", "B"),
		block("B", "A"),
		block("A"),
	}

	testza.AssertNoError(t, initializeFunctions(context.Background(), functions, "", nil, nil))
	testza.AssertEqual(t, []string{"A", "B", "C", "D"}, order)
}

func TestInitializeFunctionsUnresolvable(t *testing.T) {
	functions := []initBlock{
		noopBlock("A", "B"),
		noopBlock("B", "A"),
	}

	err := initializeFunctions(context.Background(), functions, "", nil, nil)
	testza.AssertEqual(t, "failed resolving initialization dependency tree", err.Error())
}

func TestInitializeFunctionsErrorPropagation(t *testing.T) {
	expected := errors.New("boom")

	functions := []initBlock{
		noopBlock("A"),
		{
			Name: "B",
			Func: func(_ context.Context, _ string, _ loader.AssetCache) error {
				return expected
			},
		},
	}

	err := initializeFunctions(context.Background(), functions, "", nil, nil)
	testza.AssertEqual(t, expected, errors.Cause(err))
}
