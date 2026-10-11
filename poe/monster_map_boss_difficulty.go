package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type MonsterMapBossDifficulty struct {
	raw.MonsterMapBossDifficulty
}

var MonsterMapBossDifficulties []*MonsterMapBossDifficulty

func InitializeMonsterMapBossDifficulties(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "MonsterMapBossDifficulty", &MonsterMapBossDifficulties, nil, assetCache)
}

// GetStats returns the row's stats in column order, nil where a column is empty.
func (m *MonsterMapBossDifficulty) GetStats() []*Stat {
	keys := []*int{m.Stat1, m.Stat2, m.Stat3, m.Stat4, m.Stat5}
	out := make([]*Stat, len(keys))
	for i, key := range keys {
		if key != nil {
			out[i] = Stats[*key]
		}
	}

	return out
}
