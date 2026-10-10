package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type ModType struct {
	raw.ModType
}

var ModTypes []*ModType

func InitializeModTypes(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "ModType", &ModTypes, nil, assetCache)
}

func (m *ModType) GetStats() []*Stat {
	out := make([]*Stat, len(m.Stats))
	for i, key := range m.Stats {
		out[i] = Stats[key]
	}

	return out
}
