package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type BuffTemplate struct {
	raw.BuffTemplate
}

var BuffTemplates []*BuffTemplate

var BuffTemplatesByID map[string]*BuffTemplate

func InitializeBuffTemplates(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "BuffTemplates", &BuffTemplates, func(count int64) {
		BuffTemplatesByID = make(map[string]*BuffTemplate, count)
	}, assetCache, func(obj *BuffTemplate) {
		BuffTemplatesByID[obj.ID] = obj
	})
}

func (b *BuffTemplate) GetStats() []*Stat {
	out := make([]*Stat, len(b.StatsKey))
	for i, key := range b.StatsKey {
		out[i] = Stats[key]
	}

	return out
}
