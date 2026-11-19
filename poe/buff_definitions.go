package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type BuffDefinition struct {
	raw.BuffDefinition
}

var BuffDefinitions []*BuffDefinition

var BuffDefinitionsByID map[string]*BuffDefinition

func InitializeBuffDefinitions(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "BuffDefinitions", &BuffDefinitions, func(count int64) {
		BuffDefinitionsByID = make(map[string]*BuffDefinition, count)
	}, assetCache, func(obj *BuffDefinition) {
		BuffDefinitionsByID[obj.ID] = obj
	})
}

func (b *BuffDefinition) GetStats() []*Stat {
	out := make([]*Stat, len(b.StatsKeys))
	for i, key := range b.StatsKeys {
		out[i] = Stats[key]
	}

	return out
}
