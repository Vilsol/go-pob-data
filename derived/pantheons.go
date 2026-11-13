package derived

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Vilsol/go-pob-data/poe"
)

// buildPantheons is data.pantheons of src.version.
//
//pob:block:src/Export/Scripts/pantheons.lua:9-50
func buildPantheons(src source) ([]Pantheon, error) {
	for _, init := range []func() error{
		func() error { return poe.InitializeStats(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePantheonPanelLayouts(src.ctx, src.version, src.assetCache) },
	} {
		if err := init(); err != nil {
			return nil, err
		}
	}
	describer, err := src.loadStatFile("stat_descriptions")
	if err != nil {
		return nil, err
	}

	var gods []Pantheon
	for _, p := range poe.PantheonPanelLayouts {
		if p.IsDisabled {
			continue
		}
		god := Pantheon{ID: p.ID, IsMajorGod: p.IsMajorGod, Souls: []PantheonSoul{}}
		for _, g := range []struct {
			name     string
			statKeys []int
			values   []int
		}{
			{p.GodName1, p.Effect1StatsKeys, p.Effect1Values},
			{p.GodName2, p.Effect2StatsKeys, p.Effect2Values},
			{p.GodName3, p.Effect3StatsKeys, p.Effect3Values},
			{p.GodName4, p.Effect4StatsKeys, p.Effect4Values},
		} {
			if len(g.statKeys) == 0 {
				break
			}
			soul := PantheonSoul{Name: g.name, Mods: []string{}}
			for i, key := range g.statKeys {
				if key < 0 || key >= len(poe.Stats) || i >= len(g.values) {
					return nil, fmt.Errorf("%s: stat key %d / value %d out of range", p.ID, key, i)
				}
				value := float64(g.values[i])
				lines, _, err := describer.describeStats(map[string][2]float64{poe.Stats[key].ID: {value, value}})
				if err != nil {
					return nil, err
				}
				soul.Mods = append(soul.Mods, strings.Join(lines, " "))
			}
			god.Souls = append(god.Souls, soul)
		}
		gods = append(gods, god)
	}
	if len(gods) == 0 {
		return nil, errors.New("no enabled PantheonPanelLayout row")
	}
	slices.SortFunc(gods, func(a, b Pantheon) int { return strings.Compare(a.ID, b.ID) })
	return gods, nil
}
