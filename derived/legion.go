package derived

import (
	"fmt"
	"maps"
	"math"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe"
	"github.com/Vilsol/go-pob-data/raw"
)

//pob:src/Export/Scripts/legionPassives.lua:6-56
var legionDatErrors = map[string]struct{ matchName, matchID, replaceID string }{
	"templar_notable_minimum_frenzy_charge":    {matchName: "Powerful Faith", replaceID: "templar_notable_minimum_power_charge"},
	"templar_notable_minimum_power_charge":     {matchName: "Frenzied Faith", replaceID: "templar_notable_minimum_frenzy_charge"},
	"karui_notable_add_physical_taken_as_fire": {matchID: "karui_notable_add_physical_taken_as_fire", replaceID: "karui_notable_add_rage_on_melee_hit"},
	"karui_notable_add_faster_burn":            {matchID: "karui_notable_add_faster_burn", replaceID: "karui_notable_add_faster_ignite"},
	"maraketh_notable_add_ailment_avoid":       {matchID: "maraketh_notable_add_ailment_avoid", replaceID: "maraketh_notable_add_stun_avoid"},
	"maraketh_notable_add_flask_effect":        {matchID: "maraketh_notable_add_flask_effect", replaceID: "maraketh_notable_add_alchemists_genius"},
}

//pob:src/Export/Scripts/legionPassives.lua:58-67
func fixLegionDatErrors(id, name string) string {
	e, ok := legionDatErrors[id]
	if !ok || (e.matchName != "" && name != e.matchName) || (e.matchID != "" && id != e.matchID) {
		return id
	}
	return e.replaceID
}

type legionStat struct {
	r         *edRange
	index     int
	statOrder float64
}

var (
	legionLeadWordRe  = regexp.MustCompile(`^[A-Za-z0-9]* `)
	legionLowerWordRe = regexp.MustCompile(`[a-z][A-Za-z0-9]*`)
)

//pob:src/Export/Scripts/legionPassives.lua:109-139
func parseLegionStats(describer *exportDescriber, id, dn string, statKeys []int, ranges [][2]int) (LegionPassive, error) {
	stats := map[string]*legionStat{}
	for idx, key := range statKeys {
		if key < 0 || key >= len(poe.Stats) || idx >= len(ranges) {
			return LegionPassive{}, fmt.Errorf("%s: stat %d (key %d) has no range", id, idx+1, key)
		}
		statID := poe.Stats[key].ID
		st := &legionStat{r: &edRange{min: float64(ranges[idx][0]), max: float64(ranges[idx][1])}, index: idx + 1, statOrder: math.Inf(1)}
		_, orders, err := describer.describeStats(map[string][2]float64{statID: {st.r.min, st.r.max}})
		if err != nil {
			return LegionPassive{}, fmt.Errorf("%s: %w", id, err)
		}
		if len(orders) > 0 {
			st.statOrder = orders[0]
		}
		stats[statID] = st
	}
	ranged := map[string]*edRange{}
	for statID, st := range stats {
		ranged[statID] = st.r
	}
	sd, _, err := describer.describeRanges(ranged)
	if err != nil {
		return LegionPassive{}, fmt.Errorf("%s: %w", id, err)
	}
	sorted := make([]string, 0, len(stats))
	for statID := range stats {
		sorted = append(sorted, statID)
	}
	sort.Strings(sorted)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := stats[sorted[i]], stats[sorted[j]]
		return a.statOrder < b.statOrder || a.statOrder == b.statOrder && a.index < b.index
	})

	if id == "vaal_keystone_2_v2" {
		// Immortal Ambition needs to be manually added
		sd = []string{
			"Energy Shield starts at zero",
			"Cannot Recharge or Regenerate Energy Shield",
			"Lose 5% of Energy Shield per second",
			"Life Leech effects are not removed when Unreserved Life is Filled",
			"Life Leech effects Recover Energy Shield instead while on Full Life",
		}
	}
	row := LegionPassive{ID: id, Dn: dn, Sd: append([]string{}, sd...), Rolls: []LegionRoll{}}
	for i := range row.Sd {
		if i >= len(sorted) {
			break
		}
		st := stats[sorted[i]]
		row.Rolls = append(row.Rolls, LegionRoll{Key: sorted[i], Fmt: st.r.fmt, Index: st.index, Min: st.r.min, Max: st.r.max})
	}
	return row, nil
}

// buildLegionPassives is the nodes and additions of src/Data/TimelessJewelData/LegionPassives.lua for
// src.version.
//
//pob:block:src/Export/Scripts/legionPassives.lua:141-247
func buildLegionPassives(src source) ([]LegionPassive, error) {
	var skills []*raw.AlternatePassiveSkill
	var additions []*raw.AlternatePassiveAddition
	for _, init := range []func() error{
		func() error { return poe.InitializeStats(src.ctx, src.version, src.assetCache) },
		func() error {
			return loader.InitHelper(src.ctx, src.version, "AlternatePassiveSkills", &skills, nil, src.assetCache)
		},
		func() error {
			return loader.InitHelper(src.ctx, src.version, "AlternatePassiveAdditions", &additions, nil, src.assetCache)
		},
	} {
		if err := init(); err != nil {
			return nil, err
		}
	}
	describer, err := src.loadStatFile("passive_skill_stat_descriptions")
	if err != nil {
		return nil, err
	}

	nodes, adds := map[string]LegionPassive{}, map[string]LegionPassive{}
	keystones, named := 0, 0
	for _, s := range skills {
		id := fixLegionDatErrors(s.ID, s.Name)
		row, err := parseLegionStats(describer, id, s.Name, s.StatsKeys,
			[][2]int{{s.Stat1Min, s.Stat1Max}, {s.Stat2Min, s.Stat2Max}, {s.Stat3Min, s.Stat3Max}, {s.Stat4Min, s.Stat4Max}})
		if err != nil {
			return nil, err
		}
		//pob:block:src/Export/Scripts/legionPassives.lua:162-163
		row.Keystone = slices.Contains(s.PassiveType, 4)
		nodes[id] = row
	}
	for id, row := range nodes {
		if row.Keystone {
			keystones++
		} else if id == "templar_devotion_node" || id == "eternal_small_blank" {
			named++
		}
	}
	for _, a := range additions {
		id := fixLegionDatErrors(a.ID, "")
		dn := strings.ReplaceAll(id, "_", " ")
		dn = legionLeadWordRe.ReplaceAllString(dn, "")
		dn = legionLeadWordRe.ReplaceAllString(dn, "")
		dn = legionLowerWordRe.ReplaceAllStringFunc(dn, func(w string) string { return strings.ToUpper(w[:1]) + w[1:] })
		row, err := parseLegionStats(describer, id, dn, a.StatsKeys, [][2]int{{a.Stat1Min, a.Stat1Max}, {a.Stat2Min, a.Stat2Max}})
		if err != nil {
			return nil, err
		}
		row.Addition = true
		adds[id] = row
	}
	if keystones == 0 || len(adds) == 0 || named != 2 {
		return nil, fmt.Errorf("%d keystones, %d additions, %d of 2 named nodes", keystones, len(adds), named)
	}

	rows := make([]LegionPassive, 0, len(nodes)+len(adds))
	for _, id := range slices.Sorted(maps.Keys(nodes)) {
		rows = append(rows, nodes[id])
	}
	for _, id := range slices.Sorted(maps.Keys(adds)) {
		rows = append(rows, adds[id])
	}
	return rows, nil
}
