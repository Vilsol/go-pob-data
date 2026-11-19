package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type WorldArea struct {
	raw.WorldArea
}

var WorldAreas []*WorldArea

var WorldAreasByID map[string]*WorldArea

func InitializeWorldAreas(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "WorldAreas", &WorldAreas, func(count int64) {
		WorldAreasByID = make(map[string]*WorldArea, count)
	}, assetCache, func(obj *WorldArea) {
		WorldAreasByID[obj.ID] = obj
	})
}

func (w *WorldArea) GetBosses() []*MonsterVariety {
	out := make([]*MonsterVariety, len(w.BossesMonsterVarietiesKeys))
	for i, key := range w.BossesMonsterVarietiesKeys {
		out[i] = MonsterVarieties[key]
	}

	return out
}
