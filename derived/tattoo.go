package derived

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe"
	"github.com/Vilsol/go-pob-data/raw"
)

func tattooSortSd(sd []string, descOrders map[string]float64) {
	sort.SliceStable(sd, func(i, j int) bool { return descOrders[sd[i]] < descOrders[sd[j]] })
}

//pob:src/Export/Scripts/tattooPassives.lua:50-74
func tattooPassiveStats(describer *exportDescriber, node *raw.PassiveSkill) ([]string, error) {
	values := []int64{node.Stat1, node.Stat2, node.Stat3, node.Stat4, node.Stat5}
	var sd []string
	descOrders := map[string]float64{}
	for idx, key := range node.Stats {
		if idx >= len(values) {
			return nil, fmt.Errorf("%s: stat %d has no value", node.ID, idx+1)
		}
		v := float64(values[idx])
		lines, orders, err := describer.describeStats(map[string][2]float64{poe.Stats[key].ID: {v, v}})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", node.ID, err)
		}
		for i, line := range lines {
			sd = append(sd, line)
			descOrders[line] = orders[i]
		}
	}
	tattooSortSd(sd, descOrders)
	return sd, nil
}

//pob:src/Export/Scripts/tattooPassives.lua:76-98
func tattooStats(describer *exportDescriber, o *poe.PassiveSkillOverride) ([]string, error) {
	stats := map[string][2]float64{}
	for idx, key := range o.Stats {
		if idx >= len(o.StatValues) {
			return nil, fmt.Errorf("%s: stat %d has no value", o.ID, idx+1)
		}
		v := float64(o.StatValues[idx])
		stats[poe.Stats[key].ID] = [2]float64{v, v}
	}
	lines, orders, err := describer.describeStats(stats)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", o.ID, err)
	}
	descOrders := map[string]float64{}
	for i, line := range lines {
		descOrders[line] = orders[i]
	}
	sd := append([]string{}, lines...)
	tattooSortSd(sd, descOrders)
	return sd, nil
}

// buildTattooPassives is the nodes of src/Data/TattooPassives.lua for src.version.
//
//pob:block:src/Export/Scripts/tattooPassives.lua:100-193
func buildTattooPassives(src source) ([]TattooPassive, error) {
	var passives []*raw.PassiveSkill
	for _, init := range []func() error{
		func() error { return poe.InitializeStats(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializeClientStrings(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveSkillOverrideTypes(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveOverrideLimits(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveSkillOverrides(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveSkillTattooTargetSets(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveSkillTattoos(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializeBaseItemTypes(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializeCurrencyExchanges(src.ctx, src.version, src.assetCache) },
		func() error {
			return loader.InitHelper(src.ctx, src.version, "PassiveSkills", &passives, nil, src.assetCache)
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
	clientString := func(id string) (string, error) {
		for _, s := range poe.ClientStrings {
			if s.ID == id {
				return s.Text, nil
			}
		}
		return "", fmt.Errorf("no ClientStrings row %s", id)
	}

	//pob:block:src/Export/Scripts/tattooPassives.lua:105-113
	tattooRows := map[string]*poe.PassiveSkillTattoo{}
	for _, t := range poe.PassiveSkillTattoos {
		tattooRows[poe.PassiveSkillOverrides[t.Override].ID] = t
	}

	nodes := map[string]TattooPassive{}
	for _, o := range poe.PassiveSkillOverrides {
		node := TattooPassive{ID: o.ID, OverrideType: poe.PassiveSkillOverrideTypes[o.Type].ID}
		tattooRow := tattooRows[o.ID]
		if tattooRow == nil {
			tattooRow = tattooRows["DisplayRandomKeystone"]
		}
		if tattooRow == nil {
			return nil, fmt.Errorf("%s has no PassiveSkillTattoos row and there is no DisplayRandomKeystone row", o.ID)
		}
		target := poe.PassiveSkillTattooTargetSets[tattooRow.Set]
		node.TargetType, node.TargetValue = target.Name, target.Qualifier
		node.MinimumConnected = o.RequiresAdjacent
		if o.RequiresAdjacent > 0 {
			text, err := clientString("PassiveSkillTattooAdjacentRequirementLower")
			if err != nil {
				return nil, err
			}
			node.ReminderText = []string{strings.ReplaceAll(text, "{}", strconv.Itoa(o.RequiresAdjacent))}
		}
		node.MaximumConnected = 100
		if o.MaxAdjacent > 0 {
			text, err := clientString("PassiveSkillTattooAdjacentRequirementUpper")
			if err != nil {
				return nil, err
			}
			node.ReminderText = []string{strings.ReplaceAll(text, "{}", strconv.Itoa(o.MaxAdjacent))}
			node.MaximumConnected = o.MaxAdjacent
		}
		node.ActiveEffectImage = o.PassiveBG + ".png"
		icon := o.NodeIcon
		var limitText string
		if o.Limit != nil {
			text, err := clientString("PassiveSkillTattooLimitReminder")
			if err != nil {
				return nil, err
			}
			limitText = strings.ReplaceAll(text, "{0}", poe.PassiveOverrideLimits[*o.Limit].Description)
		}
		name, rowID := o.Name, o.ID
		switch poe.PassiveSkillOverrideTypes[o.Type].ID {
		case "KeystoneTattoo":
			node.IsKeystone = true
			if o.AllocatedPassiveSkill == nil {
				return nil, fmt.Errorf("%s is a keystone tattoo without a passive", o.ID)
			}
			passive := passives[*o.AllocatedPassiveSkill]
			name, rowID, icon = passive.Name, passive.ID, passive.Icon
			if node.Sd, err = tattooPassiveStats(describer, passive); err != nil {
				return nil, err
			}
		default:
			if node.Sd, err = tattooStats(describer, o); err != nil {
				return nil, err
			}
		}
		node.Dn = name
		if name != "" && !node.IsKeystone {
			if legacy, ok := tattooLegacy(name); ok {
				node.Legacy = &legacy
			}
		}
		node.Icon = strings.TrimSuffix(icon, ".dds")
		if node.Icon != icon {
			node.Icon += ".png"
		}
		if limitText != "" {
			node.Sd = append(node.Sd, limitText)
		}
		if node.Sd == nil {
			node.Sd = []string{}
		}
		if rowID != "DisplayRandomKeystone" && !strings.Contains(name, "DNT") && !strings.Contains(name, "of the Test") {
			nodes[name] = node
		}
	}
	if len(nodes) == 0 {
		return nil, errors.New("no PassiveSkillOverrides row")
	}
	rows := make([]TattooPassive, 0, len(nodes))
	for _, name := range slices.Sorted(maps.Keys(nodes)) {
		rows = append(rows, nodes[name])
	}
	return rows, nil
}

//pob:block:src/Export/Scripts/tattooPassives.lua:179-184
func tattooLegacy(name string) (bool, bool) {
	base := slices.IndexFunc(poe.BaseItemTypes, func(b *poe.BaseItemType) bool { return b.Name == name })
	if base < 0 {
		return false, false
	}
	for _, c := range poe.CurrencyExchanges {
		if c.Item == base {
			return !c.EnabledInChallengeLeague, true
		}
	}
	return false, false
}
