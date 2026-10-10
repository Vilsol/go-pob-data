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

// GetFamilies needs InitializeModFamilies.
func (m *Mod) GetFamilies() []*ModFamily {
	out := make([]*ModFamily, len(m.Families))
	for i, key := range m.Families {
		out[i] = ModFamilies[key]
	}

	return out
}

// GetBuffTemplates returns BuffTemplate, then BuffTemplate2. It needs InitializeBuffTemplates.
func (m *Mod) GetBuffTemplates() []*BuffTemplate {
	out := make([]*BuffTemplate, 0, 1+len(m.BuffTemplate2))
	if m.BuffTemplate != nil {
		out = append(out, BuffTemplates[*m.BuffTemplate])
	}
	for _, key := range m.BuffTemplate2 {
		out = append(out, BuffTemplates[key])
	}

	return out
}
