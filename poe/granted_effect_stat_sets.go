package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type GrantedEffectStatSet struct {
	raw.GrantedEffectStatSet
}

var GrantedEffectStatSets []*GrantedEffectStatSet

func InitializeGrantedEffectStatSets(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "GrantedEffectStatSets", &GrantedEffectStatSets, nil, assetCache)
}

func (g *GrantedEffectStatSet) GetImplicitStats() []*Stat {
	if g.ImplicitStats == nil {
		return nil
	}

	out := make([]*Stat, len(g.ImplicitStats))
	for i, stat := range g.ImplicitStats {
		out[i] = Stats[stat]
	}
	return out
}

func (g *GrantedEffectStatSet) GetConstantStats() []*Stat {
	if g.ConstantStats == nil {
		return nil
	}

	out := make([]*Stat, len(g.ConstantStats))
	for i, stat := range g.ConstantStats {
		out[i] = Stats[stat]
	}
	return out
}
