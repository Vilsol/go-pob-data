package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterMapDifficulty struct {
	raw.MonsterMapDifficulty
}

var MonsterMapDifficulties []*MonsterMapDifficulty

func InitializeMonsterMapDifficulties(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterMapDifficulty", &MonsterMapDifficulties, nil, assetCache)
}

// GetStats returns the row's stats in column order, nil where a column is empty.
func (m *MonsterMapDifficulty) GetStats() []*Stat {
	keys := []*int{m.Stat1, m.Stat2, m.Stat3, m.Stat4}
	out := make([]*Stat, len(keys))
	for i, key := range keys {
		if key != nil {
			out[i] = Stats[*key]
		}
	}

	return out
}
