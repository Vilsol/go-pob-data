package poe

import (
	"context"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/raw"
)

type Mod struct {
	raw.Mod
}

var Mods []*Mod

func InitializeMods(ctx context.Context, version string, assetCache loader.AssetCache) error {
	return loader.InitHelper(ctx, version, "Mods", &Mods, nil, assetCache)
}

type ModStat struct {
	Stat *Stat
	Min  int
	Max  int
}

func (m *Mod) Stats() []ModStat {
	stats := make([]ModStat, 0, 4)

	if m.StatsKey1 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey1],
			Min:  m.Stat1Min,
			Max:  m.Stat1Max,
		})
	}

	if m.StatsKey2 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey2],
			Min:  m.Stat2Min,
			Max:  m.Stat2Max,
		})
	}

	if m.StatsKey3 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey3],
			Min:  m.Stat3Min,
			Max:  m.Stat3Max,
		})
	}

	if m.StatsKey4 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey4],
			Min:  m.Stat4Min,
			Max:  m.Stat4Max,
		})
	}

	if m.StatsKey5 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey5],
			Min:  m.Stat5Min,
			Max:  m.Stat5Max,
		})
	}

	if m.StatsKey6 != nil {
		stats = append(stats, ModStat{
			Stat: Stats[*m.StatsKey6],
			Min:  m.Stat6Min,
			Max:  m.Stat6Max,
		})
	}

	return stats
}
