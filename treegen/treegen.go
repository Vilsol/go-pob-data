package treegen

import (
	"cmp"
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
)

// mainTreeID is the PassiveSkillTrees row of the character passive tree.
const mainTreeID = "Default"

// maxListedOrbit: the export's group "orbits" only lists orbits 0-4; nodes on
// the 72-slot orbits 5 and 6 never add them.
const maxListedOrbit = 4

// grantedAttributeStats are the stats the export sums into grantedStrength/Dexterity/Intelligence (S, D, I).
var grantedAttributeStats = map[string]string{
	"base_strength":                     "S",
	"base_dexterity":                    "D",
	"base_intelligence":                 "I",
	"base_strength_and_dexterity":       "SD",
	"base_strength_and_intelligence":    "SI",
	"base_dexterity_and_intelligence":   "DI",
	"additional_strength_and_dexterity": "SD",
}

type gameData struct {
	passiveSkills  []passiveSkill
	stats          []stat
	characters     []character
	ascendancies   []ascendancy
	descendancies  []descendancy
	masteryGroups  []masteryGroup
	masteryEffects []masteryEffect
	jewelSlots     []jewelSlot
	reminderTexts  []reminderText
	blightRecipes  []blightRecipe
	blightResults  []blightResult
	blightItems    []blightItem
	baseItemTypes  []baseItemType
	trees          []passiveSkillTree
}

func loadGameData(loader fs.FS, gameVersion string) (*gameData, error) {
	d := &gameData{}
	l := &tableLoader{loader: loader, gameVersion: gameVersion}
	loadTable(l, "PassiveSkills", &d.passiveSkills)
	loadTable(l, "Stats", &d.stats)
	loadTable(l, "Characters", &d.characters)
	loadTable(l, "Ascendancy", &d.ascendancies)
	loadTable(l, "Descendancy", &d.descendancies)
	loadTable(l, "PassiveSkillMasteryGroups", &d.masteryGroups)
	loadTable(l, "PassiveSkillMasteryEffects", &d.masteryEffects)
	loadTable(l, "PassiveJewelSlots", &d.jewelSlots)
	loadTable(l, "ReminderText", &d.reminderTexts)
	loadTable(l, "BlightCraftingRecipes", &d.blightRecipes)
	loadTable(l, "BlightCraftingResults", &d.blightResults)
	loadTable(l, "BlightCraftingItems", &d.blightItems)
	loadTable(l, "BaseItemTypes", &d.baseItemTypes)
	loadTable(l, "PassiveSkillTrees", &d.trees)
	return d, l.err
}

// Generate builds the passive tree in the skilltree-export shape from the game files
// behind loader (see extractor.GetBundleLoader). gameVersion is the short version, e.g. "3.29".
func Generate(loader fs.FS, gameVersion string) (*Tree, error) {
	extractor.LoadParser(gameVersion)

	d, err := loadGameData(loader, gameVersion)
	if err != nil {
		return nil, err
	}

	treeIdx := slices.IndexFunc(d.trees, func(t passiveSkillTree) bool { return t.ID == mainTreeID })
	if treeIdx < 0 {
		return nil, fmt.Errorf("no PassiveSkillTrees row %q", mainTreeID)
	}
	graphPath := d.trees[treeIdx].PassiveSkillGraph + ".psg"

	f, err := loader.Open(gamePath(gameVersion, graphPath))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", graphPath)
	}
	graph, err := extractor.ParsePassiveSkillGraph(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", graphPath)
	}

	tree := &Tree{
		Tree:       mainTreeID,
		JewelSlots: make([]uint32, 0, len(d.jewelSlots)),
	}

	if tree.Classes, tree.Constants.Classes, err = buildClasses(d); err != nil {
		return nil, err
	}

	for _, desc := range d.descendancies {
		// Row 0 "None" is the placeholder for "no bloodline".
		if desc.ID == "None" {
			continue
		}
		a, err := buildAscendancy(desc.ID, desc.Name, desc.FlavourText, desc.RGBFlavourTextColour, desc.CoordinateRect)
		if err != nil {
			return nil, err
		}
		tree.AlternateAscendancies = append(tree.AlternateAscendancies, a)
	}

	for _, spo := range graph.SkillsPerOrbit {
		tree.Constants.SkillsPerOrbit = append(tree.Constants.SkillsPerOrbit, int(spo))
	}

	for _, slot := range d.jewelSlots {
		if slot.Slot == nil {
			return nil, errors.New("jewel slot without a passive skill")
		}
		tree.JewelSlots = append(tree.JewelSlots, d.passiveSkills[*slot.Slot].PassiveSkillGraphID)
	}

	if err := buildGraph(tree, graph, d); err != nil {
		return nil, err
	}

	return tree, nil
}

func buildClasses(d *gameData) ([]*Class, Ordered[int], error) {
	order := make([]int, len(d.characters))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int { return cmp.Compare(d.characters[a].IntegerID, d.characters[b].IntegerID) })

	classes := make([]*Class, 0, len(order))
	constants := make(Ordered[int], 0, len(order))
	for _, charIdx := range order {
		c := d.characters[charIdx]

		var ascs []ascendancy
		for _, a := range d.ascendancies {
			if a.Character != nil && *a.Character == charIdx {
				ascs = append(ascs, a)
			}
		}
		slices.SortFunc(ascs, func(a, b ascendancy) int { return cmp.Compare(a.ClassNo, b.ClassNo) })

		class := &Class{
			Name:         c.Name,
			BaseStr:      c.BaseStrength,
			BaseDex:      c.BaseDexterity,
			BaseInt:      c.BaseIntelligence,
			Ascendancies: make([]*Ascendancy, 0, len(ascs)),
		}
		for _, a := range ascs {
			asc, err := buildAscendancy(a.ID, a.Name, a.FlavourText, a.RGBFlavourTextColour, a.CoordinateRect)
			if err != nil {
				return nil, nil, err
			}
			class.Ascendancies = append(class.Ascendancies, asc)
		}
		classes = append(classes, class)

		constants = append(constants, Entry[int]{Key: c.AttrsAsID + "Class", Value: c.IntegerID})
	}

	return classes, constants, nil
}

// buildAscendancy only carries the flavour text block when there is flavour text.
func buildAscendancy(id, name, flavour, colour, rect string) (*Ascendancy, error) {
	a := &Ascendancy{ID: id, Name: name}
	if flavour == "" {
		return a, nil
	}

	a.FlavourText = strings.ReplaceAll(flavour, "\r\n", "\n")

	rgb, err := parseInts(colour, 3)
	if err != nil {
		return nil, errors.Wrapf(err, "bad flavour text colour for %s", id)
	}
	a.FlavourTextColour = fmt.Sprintf("%02x%02x%02x", rgb[0], rgb[1], rgb[2])

	r, err := parseInts(rect, 4)
	if err != nil {
		return nil, errors.Wrapf(err, "bad flavour text rect for %s", id)
	}
	a.FlavourTextRect = &Rect{X: r[0], Y: r[1], Width: r[2], Height: r[3]}

	return a, nil
}

func parseInts(s string, n int) ([]int, error) {
	parts := strings.Split(s, ",")
	if len(parts) != n {
		return nil, fmt.Errorf("want %d comma separated values, got %q", n, s)
	}
	out := make([]int, n)
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, errors.Wrapf(err, "in %q", s)
		}
		out[i] = v
	}
	return out, nil
}

func iconPath(ddsPath string) string {
	return strings.TrimSuffix(ddsPath, ".dds") + ".png"
}

func nodeID(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}

func buildGraph(tree *Tree, graph *extractor.PassiveSkillGraph, d *gameData) error {
	byGraphID := make(map[uint32]int, len(d.passiveSkills))
	for i, p := range d.passiveSkills {
		byGraphID[p.PassiveSkillGraphID] = i
	}

	jewelSlotBySkill := make(map[int]int, len(d.jewelSlots))
	for i, s := range d.jewelSlots {
		jewelSlotBySkill[*s.Slot] = i
	}

	recipeBySkill := make(map[int][]string)
	for _, r := range d.blightRecipes {
		if r.Result == nil || d.blightResults[*r.Result].PassiveSkill == nil {
			continue
		}
		skill := *d.blightResults[*r.Result].PassiveSkill
		if _, ok := recipeBySkill[skill]; ok {
			return fmt.Errorf("passive skill %d has more than one blight recipe", skill)
		}
		recipe := make([]string, 0, len(r.Items))
		for _, item := range r.Items {
			oil := d.blightItems[item].Oil
			if oil == nil {
				return fmt.Errorf("blight crafting item %d has no base item", item)
			}
			// The export names oils by their base item name without spaces, e.g. "GoldenOil".
			recipe = append(recipe, strings.ReplaceAll(d.baseItemTypes[*oil].Name, " ", ""))
		}
		recipeBySkill[skill] = recipe
	}

	in := make(map[uint32][]string)
	for _, g := range graph.Groups {
		for _, n := range g.Nodes {
			for _, out := range n.Out {
				in[out] = append(in[out], nodeID(n.ID))
			}
		}
	}

	tree.Nodes.Root = &RootNode{Out: make([]string, 0, len(graph.Roots)), In: []string{}}
	for _, r := range graph.Roots {
		tree.Nodes.Root.Out = append(tree.Nodes.Root.Out, nodeID(r))
	}

	for gi, g := range graph.Groups {
		group := &Group{
			X:       Float(g.X),
			Y:       Float(g.Y),
			IsProxy: g.IsJewelPositionRef,
			Orbits:  []int{},
			Nodes:   make([]string, 0, len(g.Nodes)),
		}
		// Background follows the largest of orbits 1-3 in the group unless the graph overrides it.
		background := int(g.BackgroundOverride)

		for _, n := range g.Nodes {
			skillIdx, ok := byGraphID[n.ID]
			if !ok {
				return fmt.Errorf("graph node %d has no PassiveSkills row", n.ID)
			}
			node, err := buildNode(d, skillIdx, jewelSlotBySkill, recipeBySkill)
			if err != nil {
				return errors.Wrapf(err, "node %d", n.ID)
			}
			node.Group = gi + 1
			node.Orbit = n.Orbit
			node.OrbitIndex = n.OrbitIndex
			node.Out = make([]string, 0, len(n.Out))
			for _, out := range n.Out {
				node.Out = append(node.Out, nodeID(out))
			}
			node.In = append([]string{}, in[n.ID]...)
			tree.Nodes.List = append(tree.Nodes.List, node)

			group.Nodes = append(group.Nodes, nodeID(n.ID))
			if orbit := int(n.Orbit); orbit <= maxListedOrbit && !slices.Contains(group.Orbits, orbit) {
				group.Orbits = append(group.Orbits, orbit)
			}
			if g.BackgroundOverride == 0 && n.Orbit >= 1 && n.Orbit <= 3 {
				background = max(background, int(n.Orbit))
			}
		}
		slices.Sort(group.Orbits)

		if background > 0 {
			group.Background = &GroupBackground{
				Image:       "PSGroupBackground" + strconv.Itoa(background),
				IsHalfImage: background == 3,
			}
		}

		tree.Groups = append(tree.Groups, group)

		// Bounds are the group centres truncated toward zero.
		x, y := int(g.X), int(g.Y)
		if gi == 0 {
			tree.MinX, tree.MinY, tree.MaxX, tree.MaxY = x, y, x, y
		}
		tree.MinX = min(tree.MinX, x)
		tree.MinY = min(tree.MinY, y)
		tree.MaxX = max(tree.MaxX, x)
		tree.MaxY = max(tree.MaxY, y)
	}

	return nil
}

func buildNode(d *gameData, skillIdx int, jewelSlotBySkill map[int]int, recipeBySkill map[int][]string) (*Node, error) {
	p := d.passiveSkills[skillIdx]
	node := &Node{
		Skill:                  p.PassiveSkillGraphID,
		Name:                   p.Name,
		Icon:                   iconPath(p.Icon),
		IsKeystone:             p.IsKeystone,
		IsNotable:              p.IsNotable,
		IsJewelSocket:          p.IsJewelSocket,
		IsBlighted:             p.IsAnointmentOnly,
		Recipe:                 recipeBySkill[skillIdx],
		IsAscendancyStart:      p.IsAscendancyStartingNode,
		IsMultipleChoice:       p.IsMultipleChoice,
		IsMultipleChoiceOption: p.IsMultipleChoiceOption,
		GrantedPassivePoints:   p.SkillPointsGranted,
		IsProxy:                p.IsProxyPassive,
		Stats:                  []string{},
	}

	if p.MasteryGroup != nil {
		mg := d.masteryGroups[*p.MasteryGroup]
		node.IsMastery = true
		node.InactiveIcon = iconPath(mg.InactiveIcon)
		node.ActiveIcon = iconPath(mg.ActiveIcon)
		node.ActiveEffectImage = iconPath(mg.ActiveEffectImage)
		for _, e := range mg.MasteryEffects {
			node.MasteryEffects = append(node.MasteryEffects, &MasteryEffect{
				Effect: d.masteryEffects[e].Hash16,
				Stats:  []string{},
			})
		}
	}

	if slotIdx, ok := jewelSlotBySkill[skillIdx]; ok {
		slot := d.jewelSlots[slotIdx]
		if slot.ClusterJewelSize != nil {
			if slot.ProxySlot == nil {
				return nil, errors.New("cluster jewel slot without a proxy")
			}
			node.ExpansionJewel = &ExpansionJewel{
				Size:  *slot.ClusterJewelSize,
				Index: slot.ClusterIndex,
				Proxy: nodeID(d.passiveSkills[*slot.ProxySlot].PassiveSkillGraphID),
			}
			if slot.ReplacesSlot != nil {
				node.ExpansionJewel.Parent = nodeID(d.passiveSkills[*d.jewelSlots[*slot.ReplacesSlot].Slot].PassiveSkillGraphID)
			}
		}
	}

	if len(p.Stats) > len(p.statValues()) {
		return nil, fmt.Errorf("%d stats but only %d stat value columns", len(p.Stats), len(p.statValues()))
	}
	for i, s := range p.Stats {
		for _, attr := range grantedAttributeStats[d.stats[s].ID] {
			switch attr {
			case 'S':
				node.GrantedStrength += p.statValues()[i]
			case 'D':
				node.GrantedDexterity += p.statValues()[i]
			case 'I':
				node.GrantedIntelligence += p.statValues()[i]
			}
		}
	}

	switch {
	case p.AscendancyKey != nil && p.DescendancyKey != nil:
		return nil, errors.New("both an ascendancy and a bloodline")
	case p.AscendancyKey != nil:
		node.AscendancyName = d.ascendancies[*p.AscendancyKey].ID
	case p.DescendancyKey != nil:
		node.AscendancyName = d.descendancies[*p.DescendancyKey].ID
		node.IsBloodline = true
	}

	for _, r := range p.ReminderStrings {
		node.ReminderText = append(node.ReminderText, d.reminderTexts[r].Text)
	}

	if p.FlavourText != "" {
		node.FlavourText = strings.Split(strings.ReplaceAll(p.FlavourText, "\r\n", "\n"), "\n")
	}

	switch len(p.Characters) {
	case 0:
	case 1:
		idx := d.characters[p.Characters[0]].IntegerID
		node.ClassStartIndex = &idx
	default:
		return nil, errors.New("more than one starting class")
	}

	return node, nil
}
