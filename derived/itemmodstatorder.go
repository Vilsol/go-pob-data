package derived

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/Vilsol/go-pob-data/poe"
)

// https://www.poewiki.net/wiki/Modifier#Domain
const (
	domainItem         = 1
	domainFlask        = 2
	domainJewel        = 10
	domainClusterJewel = 21
	domainTincture     = 34
)

// https://www.poewiki.net/wiki/Modifier#Generation_type
const genIntrinsic = 3

func modFamily(m *poe.Mod, i int) string {
	if i >= len(m.Families) {
		return ""
	}
	return poe.ModFamilies[m.Families[i]].ID
}

var itemExclusiveExcludeRe = regexp.MustCompile(`Royale|Cowards|Map|Ultimatum|UNUSED`)

//pob:block:src/Export/Scripts/mods.lua:282-292
func itemExclusive(mod *poe.Mod) bool {
	return (mod.Domain == domainItem || mod.Domain == domainFlask || mod.Domain == domainJewel || mod.Domain == domainClusterJewel || mod.Domain == domainTincture) &&
		mod.GenerationType == genIntrinsic &&
		len(mod.Families) > 0 && modFamily(mod, 0) != "AuraBonus" &&
		!strings.HasPrefix(mod.ID, "Synthesis") && !itemExclusiveExcludeRe.MatchString(mod.ID) &&
		!strings.HasPrefix(mod.ID, "MutatedUnique")
}

//pob:src/Export/statdesc.lua:431-455
func describeMod(describer *exportDescriber, m *poe.Mod) ([]string, []float64, error) {
	buffTemplateStats := map[string]bool{}
	var templates []int
	if m.BuffTemplate != nil {
		templates = append(templates, *m.BuffTemplate)
	}
	templates = append(templates, m.BuffTemplate2...)
	for _, key := range templates {
		for _, s := range poe.BuffTemplates[key].StatsKey {
			buffTemplateStats[poe.Stats[s].ID] = true
		}
	}
	stats := map[string][2]float64{}
	for i, s := range m.Stats() {
		if buffTemplateStats[s.Stat.ID] {
			continue
		}
		v := [2]float64{float64(s.Min), float64(s.Max)}
		//pob:block:src/Export/Scripts/mods.lua:68-72
		if i == 1 && m.ID == "JewelExpansionPassiveNodes" {
			// game data has 0 and 0, which means no description is generated
			v = [2]float64{2, 12}
		}
		stats[s.Stat.ID] = v
	}
	return describer.describeStats(stats)
}

//pob:src/Classes/Item.lua:64-68
func normaliseModLine(line string) string {
	line = modLineNumberRe.ReplaceAllString(line, "#")
	line = modLineRangeRe.ReplaceAllString(line, "#")
	return strings.ReplaceAll(strings.ToLower(line), "\n", " ")
}

var (
	modLineNumberRe = regexp.MustCompile(`\d+\.?\d*`)
	modLineRangeRe  = regexp.MustCompile(`\(-?#-#\)`)
)

// buildItemModStatOrder is Item.lua's uniqueModStatOrder over data.itemMods.ItemExclusive, which
// src/Export/Scripts/mods.lua writes to src/Data/ModItemExclusive.lua: each mod's lines and statOrders as
// mods.lua's writeMods describes them, then folded to the lowest statOrder of each key.
//
//pob:block:src/Classes/Item.lua:1361-1371
func buildItemModStatOrder(src source) ([]ItemModStatOrder, error) {
	for name, init := range map[string]func() error{
		"Stats":         func() error { return poe.InitializeStats(src.ctx, src.version, src.assetCache) },
		"ModFamily":     func() error { return poe.InitializeModFamilies(src.ctx, src.version, src.assetCache) },
		"BuffTemplates": func() error { return poe.InitializeBuffTemplates(src.ctx, src.version, src.assetCache) },
		"Mods":          func() error { return poe.InitializeMods(src.ctx, src.version, src.assetCache) },
	} {
		if err := init(); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	if len(poe.ModFamilies) == 0 {
		return nil, errors.New("ModFamily holds no rows")
	}
	describer, err := src.loadStatFiles("stat_descriptions", "tincture_stat_descriptions", "graft_stat_descriptions")
	if err != nil {
		return nil, err
	}

	exact, normalised := map[string]float64{}, map[string]float64{}
	for _, mod := range poe.Mods {
		if !itemExclusive(mod) {
			continue
		}
		lines, orders, err := describeMod(describer, mod)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", mod.ID, err)
		}
		//pob:block:src/Export/Scripts/mods.lua:73-74
		if len(orders) == 0 || strings.HasPrefix(lines[0], "DNT") {
			continue
		}
		for index, line := range lines {
			order := luaNumber(orders[index])
			if order <= 0 {
				return nil, fmt.Errorf("%s line %d has statOrder %v <= 0", mod.ID, index+1, order)
			}
			exactLine := strings.ReplaceAll(strings.ToLower(line), "\n", " ")
			statLine := normaliseModLine(line)
			exact[exactLine] = math.Min(orInf(exact, exactLine), order)
			normalised[statLine] = math.Min(orInf(normalised, statLine), order)
		}
	}
	if len(exact) == 0 {
		return nil, errors.New("ModItemExclusive described no lines")
	}

	keys := map[string]bool{}
	for line := range exact {
		keys[line] = true
	}
	for line := range normalised {
		keys[line] = true
	}
	rows := make([]ItemModStatOrder, 0, len(keys))
	for _, line := range slices.Sorted(maps.Keys(keys)) {
		rows = append(rows, ItemModStatOrder{Line: line, Exact: exact[line], Normalised: normalised[line]})
	}
	return rows, nil
}

// luaNumber is v as PoB reads it back from the file its export wrote it to: Lua's tostring, 14
// significant digits, which drops the error of describeRanges' repeated += 0.1.
func luaNumber(v float64) float64 {
	n, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'g', 14, 64), 64)
	return n
}

func orInf(m map[string]float64, k string) float64 {
	if v, ok := m[k]; ok {
		return v
	}
	return math.Inf(1)
}
