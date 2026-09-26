package loader

import (
	"testing"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/raw"
)

type layoutInner struct {
	Min *int
	Max *int
}

type layoutRow struct {
	Handlers map[string]string
	Name     string
	Inner    []layoutInner
	Key      int
}

// Pinned: a change to the canonical form renames every published asset.
func TestLayoutHashStable(t *testing.T) {
	testza.AssertEqual(t, "be5c474e", LayoutHash(layoutRow{}))
	testza.AssertEqual(t, LayoutHash(layoutRow{}), LayoutHash(&layoutRow{}))
	testza.AssertEqual(t, LayoutHash(raw.TranslationFile{}), LayoutHash(&raw.TranslationFile{}))
}

type wrappedRow struct {
	raw.CostType
}

// poe's row types embed raw's and decode through its DecodeMsg.
func TestLayoutHashWrapper(t *testing.T) {
	testza.AssertEqual(t, LayoutHash(raw.CostType{}), LayoutHash(&wrappedRow{}))
	testza.AssertEqual(t, RawURL[*raw.CostType]("3.29", "CostTypes"), RawURL[*wrappedRow]("3.29", "CostTypes"))
}

func TestLayoutHashChanges(t *testing.T) {
	type renamed struct {
		Handlers map[string]string
		Title    string
		Inner    []layoutInner
		Key      int
	}
	type reordered struct {
		Name     string
		Handlers map[string]string
		Inner    []layoutInner
		Key      int
	}
	type retyped struct {
		Handlers map[string]string
		Name     string
		Inner    []layoutInner
		Key      int64
	}
	type nestedInner struct {
		Min *int
		Max int
	}
	type nested struct {
		Handlers map[string]string
		Name     string
		Inner    []nestedInner
		Key      int
	}
	type appended struct {
		Handlers map[string]string
		Name     string
		Inner    []layoutInner
		Key      int
		Extra    bool
	}

	base := LayoutHash(layoutRow{})
	for name, v := range map[string]any{
		"rename":  renamed{},
		"reorder": reordered{},
		"retype":  retyped{},
		"nested":  nested{},
		"append":  appended{},
	} {
		testza.AssertNotEqual(t, base, LayoutHash(v), name)
	}
}

func TestAssetNaming(t *testing.T) {
	variant := TupleVariant(&raw.Mod{})
	testza.AssertEqual(t, "t-"+LayoutHash(raw.Mod{}), variant)
	testza.AssertEqual(t, "raw/Mods.msgpack.br", AssetPath("raw/Mods", ""))
	testza.AssertEqual(t, "raw/Mods."+variant+".msgpack.br", AssetPath("raw/Mods", variant))
	testza.AssertEqual(t, "https://data.pob.gg/data/3.29/raw/Mods."+variant+".msgpack.br", RawURL[*raw.Mod]("3.29", "Mods"))
	testza.AssertEqual(t,
		"https://data.pob.gg/data/3.29/stat_translations/en/stat_descriptions."+TupleVariant(raw.TranslationFile{})+".msgpack.br",
		TranslationURL("3.29", "en", "stat_descriptions"),
	)
}
