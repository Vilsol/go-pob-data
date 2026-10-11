package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type CraftingBenchOption struct {
	raw.CraftingBenchOption
}

var CraftingBenchOptions []*CraftingBenchOption

func InitializeCraftingBenchOptions(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CraftingBenchOptions", &CraftingBenchOptions, nil, assetCache)
}

func (c *CraftingBenchOption) GetMod() *Mod {
	if c.AddMod == nil {
		return nil
	}

	return Mods[*c.AddMod]
}

func (c *CraftingBenchOption) GetItemClasses() []*ItemClass {
	out := make([]*ItemClass, len(c.ItemClasses))
	for i, key := range c.ItemClasses {
		out[i] = ItemClasses[key]
	}

	return out
}

// GetItemCategories needs InitializeCraftingItemClassCategories.
func (c *CraftingBenchOption) GetItemCategories() []*CraftingItemClassCategory {
	out := make([]*CraftingItemClassCategory, len(c.CraftingItemClassCategories))
	for i, key := range c.CraftingItemClassCategories {
		out[i] = CraftingItemClassCategories[key]
	}

	return out
}
