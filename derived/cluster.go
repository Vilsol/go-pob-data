package derived

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/Vilsol/go-pob-data/loader"
	"github.com/Vilsol/go-pob-data/poe"
	"github.com/Vilsol/go-pob-data/raw"
)

func passiveSkillStats(node *raw.PassiveSkill) map[string][2]float64 {
	values := []int64{node.Stat1, node.Stat2, node.Stat3, node.Stat4, node.Stat5}
	stats := map[string][2]float64{}
	for i, key := range node.Stats {
		if i < len(values) {
			stats[poe.Stats[key].ID] = [2]float64{float64(values[i]), float64(values[i])}
		}
	}
	return stats
}

func pngIcon(icon string) string {
	if strings.HasSuffix(icon, "dds") {
		return strings.TrimSuffix(icon, "dds") + "png"
	}
	return icon
}

// buildClusterJewels is data.clusterJewels of src.version.
//
//pob:block:src/Export/Scripts/cluster.lua:1-91
func buildClusterJewels(src source) ([]ClusterJewels, error) {
	var passives []*raw.PassiveSkill
	for _, init := range []func() error{
		func() error { return poe.InitializeStats(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializeTags(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializeBaseItemTypes(src.ctx, src.version, src.assetCache) },
		func() error {
			return poe.InitializePassiveTreeExpansionJewelSizes(src.ctx, src.version, src.assetCache)
		},
		func() error { return poe.InitializePassiveTreeExpansionJewels(src.ctx, src.version, src.assetCache) },
		func() error { return poe.InitializePassiveTreeExpansionSkills(src.ctx, src.version, src.assetCache) },
		func() error {
			return poe.InitializePassiveTreeExpansionSpecialSkills(src.ctx, src.version, src.assetCache)
		},
		func() error { return poe.InitializePassiveJewelSlots(src.ctx, src.version, src.assetCache) },
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

	var out ClusterJewels
	for _, jewel := range poe.PassiveTreeExpansionJewels {
		size := poe.PassiveTreeExpansionJewelSizes[jewel.PassiveTreeExpansionJewelSizesKey]
		cj := ClusterJewel{
			BaseName:        poe.BaseItemTypes[jewel.BaseItemTypesKey].Name,
			Size:            size.Name,
			SizeIndex:       size.Key,
			MinNodes:        jewel.MinNodes,
			MaxNodes:        jewel.MaxNodes,
			TotalIndicies:   jewel.TotalIndices,
			SmallIndicies:   jewel.SmallIndices,
			NotableIndicies: jewel.NotableIndices,
			SocketIndicies:  jewel.SocketIndices,
			Skills:          []ClusterJewelSkill{},
		}
		for _, skill := range poe.PassiveTreeExpansionSkills {
			if skill.PassiveTreeExpansionJewelSizesKey != size.Key {
				continue
			}
			node := passives[skill.PassiveSkillsKey]
			tag := poe.Tags[skill.TagsKey].ID
			cs := ClusterJewelSkill{ID: node.ID, Name: node.Name, Icon: pngIcon(node.Icon), Tag: tag, Enchant: []string{}}
			if strings.Contains(tag, "old_do_not_use") {
				cs.Name = node.Name + " (Legacy)"
			}
			if skill.MasteryPassiveSkillsKey != nil {
				cs.MasteryIcon = pngIcon(passives[*skill.MasteryPassiveSkillsKey].Icon)
			}
			desc, _, err := describer.describeStats(passiveSkillStats(node))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", node.ID, err)
			}
			cs.Stats = append([]string{}, desc...)
			for _, line := range desc {
				cs.Enchant = append(cs.Enchant, "Added Small Passive Skills grant: "+line)
			}
			cj.Skills = append(cj.Skills, cs)
		}
		slices.SortStableFunc(cj.Skills, func(a, b ClusterJewelSkill) int { return strings.Compare(a.ID, b.ID) })
		out.Jewels = append(out.Jewels, cj)
	}
	if len(out.Jewels) == 0 {
		return nil, errors.New("no PassiveTreeExpansionJewels row")
	}
	slices.SortStableFunc(out.Jewels, func(a, b ClusterJewel) int { return strings.Compare(a.BaseName, b.BaseName) })

	notableSortOrder := map[string]int{}
	out.Keystones = []string{}
	for _, skill := range poe.PassiveTreeExpansionSpecialSkills {
		node := passives[skill.PassiveSkillsKey]
		if node.Notable {
			notableSortOrder[node.Name] = skill.StatsKey + 1
		}
		if node.Keystone {
			out.Keystones = append(out.Keystones, node.Name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(notableSortOrder)) {
		out.NotableSortOrder = append(out.NotableSortOrder, ClusterNotableOrder{Name: name, Order: notableSortOrder[name]})
	}

	for _, slot := range poe.PassiveJewelSlots {
		if slot.ClusterJewelSize == nil {
			continue
		}
		if slot.ProxySlot == nil {
			return nil, fmt.Errorf("jewel slot %d has a cluster size but no proxy", slot.Key)
		}
		out.OrbitOffsets = append(out.OrbitOffsets, ClusterOrbitOffsets{
			ProxyHash: passives[*slot.ProxySlot].Hash,
			Offsets:   slot.StartIndices[:min(3, len(slot.StartIndices))],
		})
	}
	slices.SortStableFunc(out.OrbitOffsets, func(a, b ClusterOrbitOffsets) int { return cmp.Compare(a.ProxyHash, b.ProxyHash) })

	return []ClusterJewels{out}, nil
}
