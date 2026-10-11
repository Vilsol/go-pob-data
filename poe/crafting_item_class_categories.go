package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type CraftingItemClassCategory struct {
	raw.CraftingItemClassCategory
}

var CraftingItemClassCategories []*CraftingItemClassCategory

var CraftingItemClassCategoriesByID map[string]*CraftingItemClassCategory

func InitializeCraftingItemClassCategories(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "CraftingItemClassCategories", &CraftingItemClassCategories, func(count int64) {
		CraftingItemClassCategoriesByID = make(map[string]*CraftingItemClassCategory, count)
	}, assetCache, func(obj *CraftingItemClassCategory) {
		CraftingItemClassCategoriesByID[obj.ID] = obj
	})
}

func (c *CraftingItemClassCategory) GetItemClasses() []*ItemClass {
	out := make([]*ItemClass, len(c.ItemClasses))
	for i, key := range c.ItemClasses {
		out[i] = ItemClasses[key]
	}

	return out
}
