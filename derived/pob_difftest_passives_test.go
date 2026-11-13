package derived

import (
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func TestPantheonsMatchPoB(t *testing.T) {
	type pobMod struct {
		Line string `json:"line"`
	}
	type pobSoul struct {
		Name string   `json:"name"`
		Mods []pobMod `json:"mods"`
	}
	type pobGod struct {
		Souls      []pobSoul `json:"souls"`
		IsMajorGod bool      `json:"isMajorGod"`
	}
	want := pobDecode[map[string]pobGod](t, "Pantheons.lua")
	rows, err := buildPantheons(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]pobGod{}
	for _, row := range rows {
		god := pobGod{IsMajorGod: row.IsMajorGod}
		for _, soul := range row.Souls {
			s := pobSoul{Name: soul.Name}
			for _, line := range soul.Mods {
				s.Mods = append(s.Mods, pobMod{Line: line})
			}
			god.Souls = append(god.Souls, s)
		}
		got[row.ID] = god
	}
	if len(got) != len(rows) {
		t.Errorf("%d rows, %d ids", len(rows), len(got))
	}
	diffKeys(t, "Pantheons", got, want, func(a, b pobGod) bool { return reflect.DeepEqual(a, b) })
}

func TestClusterJewelsMatchPoB(t *testing.T) {
	type pobSkill struct {
		Name        string   `json:"name"`
		Icon        string   `json:"icon"`
		MasteryIcon string   `json:"masteryIcon"`
		Tag         string   `json:"tag"`
		Stats       []string `json:"stats"`
		Enchant     []string `json:"enchant"`
	}
	type pobJewel struct {
		Skills          map[string]pobSkill `json:"skills"`
		Size            string              `json:"size"`
		SmallIndicies   []int               `json:"smallIndicies"`
		NotableIndicies []int               `json:"notableIndicies"`
		SocketIndicies  []int               `json:"socketIndicies"`
		SizeIndex       int                 `json:"sizeIndex"`
		MinNodes        int                 `json:"minNodes"`
		MaxNodes        int                 `json:"maxNodes"`
		TotalIndicies   int                 `json:"totalIndicies"`
	}
	want := pobDecode[struct {
		Jewels           map[string]pobJewel       `json:"jewels"`
		NotableSortOrder map[string]int            `json:"notableSortOrder"`
		OrbitOffsets     map[string]map[string]int `json:"orbitOffsets"`
		Keystones        []string                  `json:"keystones"`
	}](t, "ClusterJewels.lua")
	rows, err := buildClusterJewels(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("%d rows, want 1", len(rows))
	}
	data := rows[0]

	jewels := map[string]pobJewel{}
	for _, j := range data.Jewels {
		jewel := pobJewel{
			Size: j.Size, SizeIndex: j.SizeIndex, MinNodes: j.MinNodes, MaxNodes: j.MaxNodes, TotalIndicies: j.TotalIndicies,
			SmallIndicies: j.SmallIndicies, NotableIndicies: j.NotableIndicies, SocketIndicies: j.SocketIndicies,
			Skills: map[string]pobSkill{},
		}
		for _, s := range j.Skills {
			jewel.Skills[s.ID] = pobSkill{Name: s.Name, Icon: s.Icon, MasteryIcon: s.MasteryIcon, Tag: s.Tag, Stats: nilIfEmpty(s.Stats), Enchant: nilIfEmpty(s.Enchant)}
		}
		jewels[j.BaseName] = jewel
	}
	diffKeys(t, "jewels", jewels, want.Jewels, func(a, b pobJewel) bool {
		if !maps.EqualFunc(a.Skills, b.Skills, func(x, y pobSkill) bool { return reflect.DeepEqual(x, y) }) {
			return false
		}
		a.Skills, b.Skills = nil, nil
		return reflect.DeepEqual(a, b)
	})

	notable := map[string]int{}
	for _, n := range data.NotableSortOrder {
		notable[n.Name] = n.Order
	}
	diffKeys(t, "notableSortOrder", notable, want.NotableSortOrder, func(a, b int) bool { return a == b })

	orbits := map[string]map[string]int{}
	for _, o := range data.OrbitOffsets {
		row := map[string]int{}
		for i, start := range o.Offsets {
			row[fmt.Sprint(i)] = start
		}
		orbits[fmt.Sprint(o.ProxyHash)] = row
	}
	if len(orbits) != len(data.OrbitOffsets) {
		t.Errorf("%d orbit offsets, %d proxies", len(data.OrbitOffsets), len(orbits))
	}
	diffKeys(t, "orbitOffsets", orbits, want.OrbitOffsets, maps.Equal[map[string]int])

	if !slices.Equal(data.Keystones, want.Keystones) {
		t.Errorf("keystones %q, PoB %q", data.Keystones, want.Keystones)
	}
}

func nilIfEmpty[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return s
}

func TestLegionPassivesMatchPoB(t *testing.T) {
	type pobStat struct {
		Fmt   string  `json:"fmt"`
		Index int     `json:"index"`
		Min   float64 `json:"min"`
		Max   float64 `json:"max"`
	}
	type pobNode struct {
		Stats       map[string]pobStat `json:"stats"`
		ID          string             `json:"id"`
		Dn          string             `json:"dn"`
		Sd          []string           `json:"sd"`
		SortedStats []string           `json:"sortedStats"`
		Ks          bool               `json:"ks"`
	}
	want := pobDecode[struct {
		Nodes     []pobNode `json:"nodes"`
		Additions []pobNode `json:"additions"`
	}](t, filepath.Join("TimelessJewelData", "LegionPassives.lua"))
	rows, err := buildLegionPassives(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}

	// A node is compared as PoB reads it: its text, and the rolls of its first len(sd) sorted stats.
	type node struct {
		Dn    string
		Sd    []string
		Rolls []LegionRoll
		Ks    bool
	}
	fromPoB := func(nodes []pobNode, ks bool) map[string]node {
		out := map[string]node{}
		for _, n := range nodes {
			v := node{Dn: n.Dn, Sd: n.Sd, Ks: ks && n.Ks}
			for i := range min(len(n.Sd), len(n.SortedStats)) {
				s := n.Stats[n.SortedStats[i]]
				v.Rolls = append(v.Rolls, LegionRoll{Key: n.SortedStats[i], Fmt: s.Fmt, Index: s.Index, Min: s.Min, Max: s.Max})
			}
			out[n.ID] = v
		}
		return out
	}
	gotNodes, gotAdditions := map[string]node{}, map[string]node{}
	for _, row := range rows {
		v := node{Dn: row.Dn, Sd: nilIfEmpty(row.Sd), Rolls: nilIfEmpty(row.Rolls), Ks: row.Keystone}
		if row.Addition {
			gotAdditions[row.ID] = v
		} else {
			gotNodes[row.ID] = v
		}
	}
	equal := func(a, b node) bool { return reflect.DeepEqual(a, b) }
	diffKeys(t, "nodes", gotNodes, fromPoB(want.Nodes, true), equal)
	diffKeys(t, "additions", gotAdditions, fromPoB(want.Additions, false), equal)
}

func TestTattooPassivesMatchPoB(t *testing.T) {
	type pobNode struct {
		Legacy            *bool    `json:"legacy"`
		ID                string   `json:"id"`
		Dn                string   `json:"dn"`
		OverrideType      string   `json:"overrideType"`
		TargetType        string   `json:"targetType"`
		TargetValue       string   `json:"targetValue"`
		Icon              string   `json:"icon"`
		ActiveEffectImage string   `json:"activeEffectImage"`
		Sd                []string `json:"sd"`
		ReminderText      []string `json:"reminderText"`
		MinimumConnected  int      `json:"MinimumConnected"`
		MaximumConnected  int      `json:"MaximumConnected"`
		IsKeystone        bool     `json:"ks"`
	}
	want := pobDecode[struct {
		Nodes map[string]pobNode `json:"nodes"`
	}](t, "TattooPassives.lua")
	rows, err := buildTattooPassives(committedSource(t))
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]pobNode{}
	for _, row := range rows {
		n := pobNode(row)
		n.Sd = nilIfEmpty(n.Sd)
		got[row.Dn] = n
	}
	diffKeys(t, "nodes", got, want.Nodes, func(a, b pobNode) bool { return reflect.DeepEqual(a, b) })
}
