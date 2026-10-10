package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GemEffect struct {
	raw.GemEffect
}

var GemEffects []*GemEffect

var GemEffectsByID map[string]*GemEffect

func InitializeGemEffects(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GemEffects", &GemEffects, func(count int64) {
		GemEffectsByID = make(map[string]*GemEffect, count)
	}, assetCache, func(obj *GemEffect) {
		GemEffectsByID[obj.ID] = obj
	})
}

func (g *GemEffect) GetGrantedEffect() *GrantedEffect {
	return GrantedEffects[g.GrantedEffect]
}

func (g *GemEffect) GetGrantedEffect2() *GrantedEffect {
	if g.GrantedEffect2 == nil {
		return nil
	}

	return GrantedEffects[*g.GrantedEffect2]
}

func (g *GemEffect) GetGemTags() []*GemTag {
	out := make([]*GemTag, len(g.GemTags))
	for i, key := range g.GemTags {
		out[i] = GemTags[key]
	}

	return out
}
