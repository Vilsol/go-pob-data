package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GameConstant struct {
	raw.GameConstant
}

var GameConstants []*GameConstant

var GameConstantsByID map[string]*GameConstant

func InitializeGameConstants(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GameConstants", &GameConstants, func(count int64) {
		GameConstantsByID = make(map[string]*GameConstant, count)
	}, assetCache, func(obj *GameConstant) {
		GameConstantsByID[obj.ID] = obj
	})
}
