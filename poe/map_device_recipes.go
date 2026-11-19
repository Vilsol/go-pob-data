package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MapDeviceRecipe struct {
	raw.MapDeviceRecipe
}

var MapDeviceRecipes []*MapDeviceRecipe

var MapDeviceRecipesByID map[string]*MapDeviceRecipe

func InitializeMapDeviceRecipes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MapDeviceRecipes", &MapDeviceRecipes, func(count int64) {
		MapDeviceRecipesByID = make(map[string]*MapDeviceRecipe, count)
	}, assetCache, func(obj *MapDeviceRecipe) {
		MapDeviceRecipesByID[obj.ID] = obj
	})
}

func (m *MapDeviceRecipe) GetWorldArea() *WorldArea {
	if m.WorldArea == nil {
		return nil
	}

	return WorldAreas[*m.WorldArea]
}

func (m *MapDeviceRecipe) GetMods() []*Mod {
	out := make([]*Mod, len(m.Mods))
	for i, key := range m.Mods {
		out[i] = Mods[key]
	}

	return out
}
