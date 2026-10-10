package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type PassiveOverrideLimit struct {
	raw.PassiveOverrideLimit
}

var PassiveOverrideLimits []*PassiveOverrideLimit

var PassiveOverrideLimitsByID map[string]*PassiveOverrideLimit

func InitializePassiveOverrideLimits(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "PassiveOverrideLimits", &PassiveOverrideLimits, func(count int64) {
		PassiveOverrideLimitsByID = make(map[string]*PassiveOverrideLimit, count)
	}, assetCache, func(obj *PassiveOverrideLimit) {
		PassiveOverrideLimitsByID[obj.ID] = obj
	})
}
