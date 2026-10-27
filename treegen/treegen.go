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

// The values below are hardcoded in the game client; the game files don't contain them.
const (
	// pssCentreInnerRadius is a game-client value absent from the game files.
	pssCentreInnerRadius = 130
	// totalPoints is a game-client value absent from the game files.
	totalPoints = 123
	// ascendancyPoints is a game-client value absent from the game files.
	ascendancyPoints = 8
)

// orbitRadii is a game-client value absent from the game files, one radius per orbit.
var orbitRadii = []int{0, 82, 162, 335, 493, 662, 846}

// characterAttributes is a game-client value absent from the game files.
var characterAttributes = Ordered[int]{
	{Key: "Strength", Value: 0},
	{Key: "Dexterity", Value: 1},
	{Key: "Intelligence", Value: 2},
}

// grantedAttributeStats are the stats the export sums into grantedStrength/Dexterity/Intelligence (S, D, I).
var grantedAttributeStats = map[string]string{
	"base_strength":                     "S",
	"base_dexterity":                    "D",
	"base_intelligence":                 "I",
	"base_strength_and_dexterity":       "SD",
	"base_strength_and_intelligence":    "SI",
	"base_dexterity_and_intelligence":   "DI",
	"additional_strength_and_dexterity": "SD",
	"additional_all_attributes":         "SDI",
}

type gameData struct {
	descriptions           *descriptions
	reminderByID           map[string]string
	passiveSkills          []passiveSkill
	stats                  []stat
	characters             []character
	ascendancies           []ascendancy
	descendancies          []descendancy
	masteryGroups          []masteryGroup
	masteryEffects         []masteryEffect
	jewelSlots             []jewelSlot
	reminderTexts          []reminderText
	blightRecipes          []blightRecipe
	blightResults          []blightResult
	blightItems            []blightItem
	baseItemTypes          []baseItemType
	trees                  []passiveSkillTree
	expansionSkills        []expansionSkill
	expansionSpecialSkills []expansionSpecialSkill
	overrides              []passiveSkillOverride
	buffTemplates          []buffTemplate
	buffDefinitions        []buffDefinition
	clientStrings          []clientString
}

func loadGameData(loader fs.FS, gameVersion string) (*gameData, error) {
	d := &gameData{}
	l := &tableLoader{loader: loader, game: extractor.PoE1, gameVersion: gameVersion}
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
	loadTable(l, "PassiveTreeExpansionSkills", &d.expansionSkills)
	loadTable(l, "PassiveTreeExpansionSpecialSkills", &d.expansionSpecialSkills)
	loadTable(l, "PassiveSkillOverrides", &d.overrides)
	loadTable(l, "BuffTemplates", &d.buffTemplates)
	loadTable(l, "BuffDefinitions", &d.buffDefinitions)
	loadTable(l, "ClientStrings", &d.clientStrings)
	if l.err != nil {
		return nil, l.err
	}

	d.reminderByID = make(map[string]string, len(d.reminderTexts))
	for _, r := range d.reminderTexts {
		d.reminderByID[r.ID] = r.Text
	}

	var err error
	d.descriptions, err = loadDescriptions(loader, gameVersion)
	return d, err
}

// Generate builds the passive tree in the skilltree-export shape from the game files
// behind loader (see extractor.GetBundleLoader). gameVersion is the short version, e.g. "3.29".
func Generate(loader fs.FS, gameVersion string) (*Tree, error) {
	tree, _, err := generate(loader, gameVersion)
	return tree, err
}

func generate(loader fs.FS, gameVersion string) (*Tree, *gameData, error) {
	extractor.LoadParser(extractor.PoE1, gameVersion)

	d, err := loadGameData(loader, gameVersion)
	if err != nil {
		return nil, nil, err
	}

	graph, err := loadGraph(loader, gameVersion, d, mainTreeID)
	if err != nil {
		return nil, nil, err
	}

	tree := &Tree{
		Tree:       mainTreeID,
		JewelSlots: make([]uint32, 0, len(d.jewelSlots)),
		Constants: Constants{
			CharacterAttributes:  characterAttributes,
			PSSCentreInnerRadius: pssCentreInnerRadius,
			OrbitRadii:           orbitRadii,
		},
		Points: Points{TotalPoints: totalPoints, AscendancyPoints: ascendancyPoints},
	}

	if tree.Classes, tree.Constants.Classes, err = buildClasses(d); err != nil {
		return nil, nil, err
	}

	for _, desc := range d.descendancies {
		// Row 0 "None" is the placeholder for "no bloodline".
		if desc.ID == "None" {
			continue
		}
		a, err := buildAscendancy(desc.ID, desc.Name, desc.FlavourText, desc.RGBFlavourTextColour, desc.CoordinateRect)
		if err != nil {
			return nil, nil, err
		}
		tree.AlternateAscendancies = append(tree.AlternateAscendancies, a)
	}

	for _, spo := range graph.SkillsPerOrbit {
		tree.Constants.SkillsPerOrbit = append(tree.Constants.SkillsPerOrbit, int(spo))
	}
	if len(graph.SkillsPerOrbit) != len(orbitRadii) {
		return nil, nil, fmt.Errorf("graph has %d orbits but orbitRadii has %d", len(graph.SkillsPerOrbit), len(orbitRadii))
	}

	for _, slot := range d.jewelSlots {
		if slot.Slot == nil {
			return nil, nil, errors.New("jewel slot without a passive skill")
		}
		tree.JewelSlots = append(tree.JewelSlots, d.passiveSkills[*slot.Slot].PassiveSkillGraphID)
	}

	lookups, err := buildLookups(d)
	if err != nil {
		return nil, nil, err
	}

	if err := buildGraph(tree, graph, d, lookups); err != nil {
		return nil, nil, err
	}

	if err := buildClusterNodes(tree, d, lookups); err != nil {
		return nil, nil, err
	}

	if err := buildGrantedNodes(tree, d, lookups); err != nil {
		return nil, nil, err
	}

	return tree, d, nil
}

// loadGraph parses the passive skill graph of a PassiveSkillTrees row.
func loadGraph(loader fs.FS, gameVersion string, d *gameData, treeID string) (*extractor.PassiveSkillGraph, error) {
	treeIdx := slices.IndexFunc(d.trees, func(t passiveSkillTree) bool { return t.ID == treeID })
	if treeIdx < 0 {
		return nil, fmt.Errorf("no PassiveSkillTrees row %q", treeID)
	}
	graphPath := d.trees[treeIdx].PassiveSkillGraph + ".psg"

	f, err := loader.Open(gamePath(gameVersion, graphPath))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", graphPath)
	}
	graph, err := extractor.ParsePassiveSkillGraph(f, extractor.PoE1)
	return graph, errors.Wrapf(err, "failed to parse %s", graphPath)
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
	if ddsPath == "" {
		return ""
	}
	return strings.TrimSuffix(ddsPath, ".dds") + ".png"
}

func nodeID(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}

// lookups index game data by PassiveSkills row.
type lookups struct {
	jewelSlotBySkill map[int]int
	recipeBySkill    map[int][]string
}

func buildLookups(d *gameData) (*lookups, error) {
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
			return nil, fmt.Errorf("passive skill %d has more than one blight recipe", skill)
		}
		recipe := make([]string, 0, len(r.Items))
		for _, item := range r.Items {
			oil := d.blightItems[item].Oil
			if oil == nil {
				return nil, fmt.Errorf("blight crafting item %d has no base item", item)
			}
			// The export names oils by their base item name without spaces, e.g. "GoldenOil".
			recipe = append(recipe, strings.ReplaceAll(d.baseItemTypes[*oil].Name, " ", ""))
		}
		recipeBySkill[skill] = recipe
	}

	return &lookups{jewelSlotBySkill: jewelSlotBySkill, recipeBySkill: recipeBySkill}, nil
}

func buildGraph(tree *Tree, graph *extractor.PassiveSkillGraph, d *gameData, l *lookups) error {
	byGraphID := make(map[uint32]int, len(d.passiveSkills))
	for i, p := range d.passiveSkills {
		byGraphID[p.PassiveSkillGraphID] = i
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
			node, err := buildNode(d, skillIdx, l)
			if err != nil {
				return errors.Wrapf(err, "node %d", n.ID)
			}
			node.Position = &Position{
				Group:      gi + 1,
				Orbit:      n.Orbit,
				OrbitIndex: n.OrbitIndex,
				Out:        make([]string, 0, len(n.Out)),
				In:         append([]string{}, in[n.ID]...),
			}
			for _, out := range n.Out {
				node.Out = append(node.Out, nodeID(out))
			}
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

// buildClusterNodes adds the nodes cluster jewels place on the tree, in the export's order:
// the special (notable/keystone) skills, then each small skill followed by its mastery.
func buildClusterNodes(tree *Tree, d *gameData, l *lookups) error {
	rows := make([]int, 0, len(d.expansionSpecialSkills)+2*len(d.expansionSkills))
	for _, s := range d.expansionSpecialSkills {
		rows = append(rows, s.PassiveSkill)
	}
	masteries := make(map[int]bool)
	for _, s := range d.expansionSkills {
		rows = append(rows, s.PassiveSkill)
		if s.Mastery != nil {
			rows = append(rows, *s.Mastery)
			masteries[*s.Mastery] = true
		}
	}
	if len(masteries) == 0 {
		return errors.New("no cluster jewel mastery passives")
	}

	seen := make(map[int]bool, len(rows))
	for _, row := range rows {
		if seen[row] {
			continue
		}
		seen[row] = true

		node, err := buildNode(d, row, l)
		if err != nil {
			return errors.Wrapf(err, "cluster node %d", d.passiveSkills[row].PassiveSkillGraphID)
		}
		node.IsMastery = node.IsMastery || masteries[row]
		tree.Nodes.List = append(tree.Nodes.List, node)
	}
	return nil
}

// forbiddenJewelPrefix marks the ascendancy passives only Forbidden Flame/Flesh grant: GGG names
// them AscendancySpecialEldritch*, and the jewels' mods roll the passive hash at runtime.
const forbiddenJewelPrefix = "AscendancySpecial"

// buildGrantedNodes adds the passives an item can allocate that are neither in the graph nor on
// a cluster jewel: the ascendancy passives in the Forbidden Tattoo pools (PassiveSkillOverrides)
// and the Forbidden Flame/Flesh ascendancy passives.
func buildGrantedNodes(tree *Tree, d *gameData, l *lookups) error {
	emitted := make(map[uint32]bool, len(tree.Nodes.List))
	for _, n := range tree.Nodes.List {
		emitted[n.Skill] = true
	}

	granted := make(map[int]bool)
	for _, o := range d.overrides {
		if o.AllocatedPassiveSkill != nil {
			granted[*o.AllocatedPassiveSkill] = true
		}
	}
	if len(granted) == 0 {
		return errors.New("no PassiveSkillOverrides allocate a passive")
	}

	forbidden := 0
	for i, p := range d.passiveSkills {
		if p.AscendancyKey != nil && strings.HasPrefix(p.ID, forbiddenJewelPrefix) {
			granted[i] = true
			forbidden++
		}
	}
	if forbidden == 0 {
		return fmt.Errorf("no %s ascendancy passives", forbiddenJewelPrefix)
	}

	rows := make([]int, 0, len(granted))
	for row := range granted {
		if !emitted[d.passiveSkills[row].PassiveSkillGraphID] {
			rows = append(rows, row)
		}
	}
	slices.Sort(rows)

	for _, row := range rows {
		node, err := buildNode(d, row, l)
		if err != nil {
			return errors.Wrapf(err, "granted node %d", d.passiveSkills[row].PassiveSkillGraphID)
		}
		tree.Nodes.List = append(tree.Nodes.List, node)
	}
	return nil
}

func buildNode(d *gameData, skillIdx int, l *lookups) (*Node, error) {
	p := d.passiveSkills[skillIdx]
	node := &Node{
		Skill:                  p.PassiveSkillGraphID,
		Name:                   p.Name,
		Icon:                   iconPath(p.Icon),
		IsKeystone:             p.IsKeystone,
		IsNotable:              p.IsNotable,
		IsJewelSocket:          p.IsJewelSocket,
		IsBlighted:             p.IsAnointmentOnly,
		Recipe:                 l.recipeBySkill[skillIdx],
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
			effect, err := d.describeMasteryEffect(&d.masteryEffects[e])
			if err != nil {
				return nil, errors.Wrapf(err, "mastery effect %d", d.masteryEffects[e].Hash16)
			}
			node.MasteryEffects = append(node.MasteryEffects, effect)
		}
	}

	if slotIdx, ok := l.jewelSlotBySkill[skillIdx]; ok {
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

	if err := d.describeNode(node, &p); err != nil {
		return nil, err
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
