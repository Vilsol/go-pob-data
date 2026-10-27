package treegen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/extractor"
	st "github.com/Vilsol/go-pob-data/stat_translations"
)

// ignoreFilter2 is passivetree.lua:15's "^%[DNT".
var ignoreFilter2 = regexp.MustCompile(`^\[DNT`)

// dntKept2 are the [DNT] names passivetree.lua:687 keeps.
var dntKept2 = []string{"[DNT] Kite Fisher", "[DNT] Troller", "[DNT] Spearfisher", "[DNT] Angler", "[DNT] Whaler"}

// baseAttributes2 are the hardcoded attribute nodes of passivetree.lua:613-617, in its order.
var baseAttributes2 = []uint32{26297, 14927, 57022}

// orbitRadii2 is passivetree.lua:470-472, a game-client value.
var orbitRadii2 = []float64{0, 82, 162, 335, 493, 662, 846, 251, 1080, 1322}

// arcAngle2 is passivetree.lua:1118, by ascendancy count.
var arcAngle2 = map[int]float64{0: 0, 1: 0, 2: 12, 3: 24, 4: 36, 5: 48, 6: 60, 7: 72, 8: 84, 9: 96}

const (
	poe2StatDescriptions = "Data/StatDescriptions/passive_skill_stat_descriptions.csd"
	poe2UIImages         = "Art/UIImages1.txt"
)

type Tree2 struct {
	Nodes         map[string]*Node2  `json:"nodes"`
	NodeOverlay   map[string]Overlay `json:"nodeOverlay"`
	ConnectionArt map[string]string  `json:"connectionArt"`
	Tree          string             `json:"tree"`
	Classes       []*Class2          `json:"classes"`
	Groups        []*Group2          `json:"groups"`
	JewelSlots    []uint32           `json:"jewelSlots"`
	Constants     Constants2         `json:"constants"`
	MinX          float64            `json:"min_x"`
	MinY          float64            `json:"min_y"`
	MaxX          float64            `json:"max_x"`
	MaxY          float64            `json:"max_y"`
}

type Constants2 struct {
	Classes              map[string]int `json:"classes"`
	CharacterAttributes  map[string]int `json:"characterAttributes"`
	SkillsPerOrbit       []int          `json:"skillsPerOrbit"`
	OrbitAnglesByOrbit   [][]float64    `json:"orbitAnglesByOrbit"`
	OrbitRadii           []float64      `json:"orbitRadii"`
	PSSCentreInnerRadius int            `json:"PSSCentreInnerRadius"`
}

type Class2 struct {
	Name         string         `json:"name"`
	Ascendancies []*Ascendancy2 `json:"ascendancies"`
	Background   Background2    `json:"background"`
	IntegerID    int            `json:"integerId"`
	BaseStr      int            `json:"base_str"`
	BaseDex      int            `json:"base_dex"`
	BaseInt      int            `json:"base_int"`
}

type Ascendancy2 struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	InternalID string      `json:"internalId"`
	Replace    string      `json:"replace,omitempty"`
	ReplaceBy  string      `json:"replaceBy,omitempty"`
	Background Background2 `json:"background"`
}

type Background2 struct {
	Active  *Size2  `json:"active,omitempty"`
	Bg      *Size2  `json:"bg,omitempty"`
	Image   string  `json:"image"`
	Section string  `json:"section"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
}

type Size2 struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Group2 struct {
	Orbits []int    `json:"orbits"`
	Nodes  []uint32 `json:"nodes"`
	X      float64  `json:"x"`
	Y      float64  `json:"y"`
}

type Connection2 struct {
	ID    uint32 `json:"id"`
	Orbit int32  `json:"orbit"`
}

type Overlay struct {
	Alloc   string `json:"alloc"`
	Path    string `json:"path"`
	Unalloc string `json:"unalloc"`
}

type UnlockConstraint2 struct {
	Ascendancy string   `json:"ascendancy,omitempty"`
	Nodes      []uint32 `json:"nodes"`
}

// NodeOption2 is one entry of a node's options: a base attribute, a class override or an
// ascendancy replacement.
type NodeOption2 struct {
	NodeOverlay    *Overlay `json:"nodeOverlay,omitempty"`
	ID             *uint32  `json:"id,omitempty"`
	Name           string   `json:"name,omitempty"`
	Icon           string   `json:"icon,omitempty"`
	AscendancyName string   `json:"ascendancyName,omitempty"`
	Stats          []string `json:"stats,omitempty"`
}

//nolint:govet // fields follow tree.json's key order
type Node2 struct {
	Skill                  uint32             `json:"skill"`
	Group                  int                `json:"group"`
	Orbit                  uint32             `json:"orbit"`
	OrbitIndex             uint32             `json:"orbitIndex"`
	Connections            []Connection2      `json:"connections"`
	Name                   *string            `json:"name,omitempty"`
	StringID               string             `json:"stringId,omitempty"`
	Icon                   *string            `json:"icon,omitempty"`
	FlavourText            string             `json:"flavourText,omitempty"`
	IsKeystone             bool               `json:"isKeystone,omitempty"`
	IsNotable              bool               `json:"isNotable,omitempty"`
	IsOnlyImage            bool               `json:"isOnlyImage,omitempty"`
	IsJewelSocket          bool               `json:"isJewelSocket,omitempty"`
	AliasPassiveSocket     string             `json:"aliasPassiveSocket,omitempty"`
	Sinister               bool               `json:"sinister,omitempty"`
	NoRadius               bool               `json:"noRadius,omitempty"`
	AscendancyName         string             `json:"ascendancyName,omitempty"`
	IsAscendancyStart      bool               `json:"isAscendancyStart,omitempty"`
	ContainJewelSocket     bool               `json:"containJewelSocket,omitempty"`
	NodeOverlay            *Overlay           `json:"nodeOverlay,omitempty"`
	UnlockConstraint       *UnlockConstraint2 `json:"unlockConstraint,omitempty"`
	ConnectionArt          string             `json:"connectionArt,omitempty"`
	Stats                  []string           `json:"stats"`
	ActiveEffectImage      string             `json:"activeEffectImage,omitempty"`
	IsAttribute            bool               `json:"isAttribute,omitempty"`
	Options                any                `json:"options,omitempty"`
	Recipe                 []string           `json:"recipe,omitempty"`
	IsSwitchable           bool               `json:"isSwitchable,omitempty"`
	ClassesStart           []string           `json:"classesStart,omitempty"`
	IsMultipleChoice       bool               `json:"isMultipleChoice,omitempty"`
	IsMultipleChoiceOption bool               `json:"isMultipleChoiceOption,omitempty"`
	IsFreeAllocate         bool               `json:"isFreeAllocate,omitempty"`
	ApplyToArmour          bool               `json:"applyToArmour,omitempty"`
}

type gameData2 struct {
	describer           *st.TranslationParser
	uiImages            map[string]string
	bySkillID           map[uint32]int
	passiveSkills       []passiveSkill2
	stats               []stat
	characters          []character2
	ascendancies        []ascendancy2
	uiArt               []treeUIArt2
	frameArt            []nodeFrameArt2
	masteryGroups       []masteryGroup2
	masteryArt          []masteryArt2
	classOverrides      []classOverride2
	ascendancyOverrides []ascendancyOverride2
	blightResults       []blightResult2
	blightRecipes       []blightRecipe2
	blightItems         []blightItem2
	skillGems           []skillGem2
	baseItemTypes       []baseItemType
	jewelSlots          []jewelSlot
	trees               []passiveSkillTree
}

func loadGameData2(loader fs.FS, gameVersion string) (*gameData2, error) {
	d := &gameData2{}
	l := &tableLoader{loader: loader, game: extractor.PoE2, gameVersion: gameVersion}
	loadTable(l, "PassiveSkills", &d.passiveSkills)
	loadTable(l, "Stats", &d.stats)
	loadTable(l, "Characters", &d.characters)
	loadTable(l, "Ascendancy", &d.ascendancies)
	loadTable(l, "PassiveSkillTreeUIArt", &d.uiArt)
	loadTable(l, "PassiveSkillTreeNodeFrameArt", &d.frameArt)
	loadTable(l, "PassiveSkillMasteryGroups", &d.masteryGroups)
	loadTable(l, "PassiveSkillTreeMasteryArt", &d.masteryArt)
	loadTable(l, "ClassPassiveSkillOverrides", &d.classOverrides)
	loadTable(l, "AscendancyPassiveSkillOverrides", &d.ascendancyOverrides)
	loadTable(l, "BlightCraftingResults", &d.blightResults)
	loadTable(l, "BlightCraftingRecipes", &d.blightRecipes)
	loadTable(l, "BlightCraftingItems", &d.blightItems)
	loadTable(l, "SkillGems", &d.skillGems)
	loadTable(l, "BaseItemTypes", &d.baseItemTypes)
	loadTable(l, "PassiveJewelSlots", &d.jewelSlots)
	loadTable(l, "PassiveSkillTrees", &d.trees)
	if l.err != nil {
		return nil, l.err
	}

	// GetRow("PassiveSkillNodeId", id) is the first row with that id.
	d.bySkillID = make(map[uint32]int, len(d.passiveSkills))
	for i := len(d.passiveSkills) - 1; i >= 0; i-- {
		d.bySkillID[d.passiveSkills[i].PassiveSkillGraphID] = i
	}

	var err error
	d.describer, err = st.NewDescriber(loader, gameFilePath(extractor.PoE2, gameVersion, poe2StatDescriptions))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load %s", poe2StatDescriptions)
	}

	d.uiImages, err = loadUIImages(loader, gameFilePath(extractor.PoE2, gameVersion, poe2UIImages))
	return d, err
}

// loadUIImages maps each lowercased image name of art/uiimages1.txt to its lowercased path
// (AssetSheets.lua:123-153).
func loadUIImages(loader fs.FS, name string) (map[string]string, error) {
	f, err := loader.Open(name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", name)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", name)
	}
	b = bytes.TrimPrefix(b, []byte{0xff, 0xfe})
	if len(b)%2 != 0 {
		return nil, fmt.Errorf("%s is not UTF-16", name)
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = uint16(b[2*i]) | uint16(b[2*i+1])<<8
	}

	images := make(map[string]string)
	for _, line := range strings.FieldsFunc(string(utf16.Decode(u)), func(r rune) bool { return r == '\r' || r == '\n' }) {
		fields := strings.Fields(strings.ReplaceAll(line, `"`, " "))
		if len(fields) < 2 {
			continue
		}
		images[strings.ToLower(fields[0])] = strings.ToLower(fields[1])
	}
	return images, nil
}

func (d *gameData2) uiImage(name string) (string, error) {
	p, ok := d.uiImages[strings.ToLower(name)]
	if !ok {
		return "", fmt.Errorf("no ui image %q", name)
	}
	return p, nil
}

func (d *gameData2) frameOverlay(frame int) (*Overlay, error) {
	art := d.frameArt[frame]
	active, err := d.uiImage(art.Active)
	if err != nil {
		return nil, err
	}
	canAllocate, err := d.uiImage(art.CanAllocate)
	if err != nil {
		return nil, err
	}
	normal, err := d.uiImage(art.Normal)
	if err != nil {
		return nil, err
	}
	return &Overlay{Alloc: active, Path: canAllocate, Unalloc: normal}, nil
}

func ascendancyOverlay(name string, notable bool) *Overlay {
	typeFrame := "Small"
	if notable {
		typeFrame = "Large"
	}
	return &Overlay{
		Alloc:   name + "Frame" + typeFrame + "Allocated",
		Path:    name + "Frame" + typeFrame + "CanAllocate",
		Unalloc: name + "Frame" + typeFrame + "Normal",
	}
}

func (d *gameData2) describe(p *passiveSkill2) ([]string, error) {
	ids := make([]string, len(p.Stats))
	for i, s := range p.Stats {
		ids[i] = d.stats[s].ID
	}
	values := p.statValues()
	if len(ids) > len(values) {
		return nil, fmt.Errorf("passive %s has %d stats", p.ID, len(ids))
	}
	described, err := d.describer.Describe(ids, values[:len(ids)])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to describe %s", p.ID)
	}
	// statdesc.lua:49 strips GGG markup from the wording and :410 splits it at its line breaks.
	var lines []string
	for _, line := range described.Lines {
		for _, l := range strings.Split(strings.ReplaceAll(escapeGGGString(line), `\n`, "\n"), "\n") {
			lines = append(lines, sanitiseText(l))
		}
	}
	return lines, nil
}

// escapeGGGString is Common.lua:1078-1084.
var (
	gggMarkup  = regexp.MustCompile(`<[^>]+>\{([^}]+)\}`)
	gggLink    = regexp.MustCompile(`\[([^|\]]+)\]`)
	gggAltLink = regexp.MustCompile(`\[[^|]+\|([^|]+)\]`)
)

func escapeGGGString(s string) string {
	s = gggMarkup.ReplaceAllString(s, "$1")
	s = gggLink.ReplaceAllString(s, "$1")
	return gggAltLink.ReplaceAllString(s, "$1")
}

// sanitiseText is Common.lua:321-347; like the Lua it works on bytes, so any byte past 127 that
// survives the named replacements becomes "?".
var (
	sanitiseTags  = regexp.MustCompile(`<[^<>]*>`)
	sanitiseNamed = strings.NewReplacer(
		"\u2010", "-", "\u2011", "-", "\u2012", "-", "\u2013", "-", "\u2014", "-", "\u2015", "-", "\u2212", "-",
		"\u2022 ", "", "\u2022", "",
		"\u00e4", "a", "\u00f6", "o", "\u00ed", "i", "\u00f3", "o",
	)
	sanitiseBytes = strings.NewReplacer("\x96", "-", "\x97", "-", "\xe4", "a", "\xf6", "o")
)

func sanitiseText(text string) string {
	if !strings.ContainsFunc(text, func(r rune) bool { return r >= 0x80 || r == '<' }) {
		return text
	}
	for sanitiseTags.MatchString(text) {
		text = sanitiseTags.ReplaceAllString(text, "")
	}
	b := []byte(sanitiseBytes.Replace(sanitiseNamed.Replace(text)))
	for i, c := range b {
		if c >= 0x80 {
			b[i] = '?'
		}
	}
	return string(b)
}

// roundTo is passivetree.lua:82-85.
func roundTo(num float64, places int) float64 {
	m := math.Pow(10, float64(places))
	return math.Floor(num*m+0.5) / m
}

// calcOrbitAngles is passivetree.lua:22-43.
func calcOrbitAngles(nodesInOrbit int) []float64 {
	var degrees []float64
	switch nodesInOrbit {
	case 16:
		degrees = []float64{0, 30, 45, 60, 90, 120, 135, 150, 180, 210, 225, 240, 270, 300, 315, 330}
	case 40:
		degrees = []float64{0, 10, 20, 30, 40, 45, 50, 60, 70, 80, 90, 100, 110, 120, 130, 135, 140, 150, 160, 170, 180, 190, 200, 210, 220, 225, 230, 240, 250, 260, 270, 280, 290, 300, 310, 315, 320, 330, 340, 350}
	default:
		for i := 0; i <= nodesInOrbit; i++ {
			degrees = append(degrees, 360*float64(i)/float64(nodesInOrbit))
		}
	}
	out := make([]float64, len(degrees))
	for i, deg := range degrees {
		out[i] = deg * math.Pi / 180
	}
	return out
}

// GeneratePoE2 builds the PoE2 passive tree from the game files behind loader in the shape of
// PathOfBuilding-PoE2's src/TreeData/<version>/tree.json, as its src/Export/Scripts/passivetree.lua
// does, without the sprite sheets (assets, ddsCoords). gameVersion is the short version, e.g. "0.5".
func GeneratePoE2(loader fs.FS, gameVersion string) (*Tree2, error) {
	extractor.LoadParser(extractor.PoE2, gameVersion)

	d, err := loadGameData2(loader, gameVersion)
	if err != nil {
		return nil, err
	}

	treeIdx := slices.IndexFunc(d.trees, func(t passiveSkillTree) bool { return t.ID == mainTreeID })
	if treeIdx < 0 {
		return nil, fmt.Errorf("no PassiveSkillTrees row %q", mainTreeID)
	}
	graphPath := d.trees[treeIdx].PassiveSkillGraph + ".psg"
	f, err := loader.Open(gameFilePath(extractor.PoE2, gameVersion, graphPath))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", graphPath)
	}
	graph, err := extractor.ParsePassiveSkillGraph(f, extractor.PoE2)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", graphPath)
	}

	return buildTree2(d, graph)
}

//nolint:gocognit,gocyclo,cyclop,maintidx // a port of passivetree.lua's generation, kept in its order
func buildTree2(d *gameData2, graph *extractor.PassiveSkillGraph) (*Tree2, error) {
	tree := &Tree2{
		Tree:  mainTreeID,
		Nodes: make(map[string]*Node2),
		Constants: Constants2{
			Classes: map[string]int{
				"StrDexIntClass": 0, "StrClass": 1, "DexClass": 2, "IntClass": 3,
				"StrDexClass": 4, "StrIntClass": 5, "DexIntClass": 6,
			},
			CharacterAttributes:  map[string]int{"Strength": 0, "Dexterity": 1, "Intelligence": 2},
			PSSCentreInnerRadius: pssCentreInnerRadius,
			OrbitRadii:           orbitRadii2,
		},
		NodeOverlay: map[string]Overlay{
			"Normal":   {Alloc: "PSSkillFrameActive", Path: "PSSkillFrameHighlighted", Unalloc: "PSSkillFrame"},
			"Notable":  {Alloc: "NotableFrameAllocated", Path: "NotableFrameCanAllocate", Unalloc: "NotableFrameUnallocated"},
			"Keystone": {Alloc: "KeystoneFrameAllocated", Path: "KeystoneFrameCanAllocate", Unalloc: "KeystoneFrameUnallocated"},
			"Socket":   {Alloc: "JewelFrameAllocated", Path: "JewelFrameCanAllocate", Unalloc: "JewelFrameUnallocated"},
		},
		ConnectionArt: map[string]string{"default": "Character", "ascendancy": "CharacterAscendancy"},
	}

	row := func(id uint32) *passiveSkill2 {
		if i, ok := d.bySkillID[id]; ok {
			return &d.passiveSkills[i]
		}
		return nil
	}

	// passivetree.lua:502-609
	ascendancyReplacements := make(map[string]string)
	var replacementOrder []string
	for _, classID := range graph.Roots {
		p := row(classID)
		if p == nil {
			return nil, fmt.Errorf("class start %d has no passive skill", classID)
		}
		if ignoreFilter2.MatchString(p.Name) {
			continue
		}
		for _, charIdx := range p.Characters {
			c := d.characters[charIdx]
			if ignoreFilter2.MatchString(c.Name) {
				continue
			}
			class := &Class2{
				Name:         c.Name,
				IntegerID:    c.IntegerID,
				BaseStr:      c.BaseStrength,
				BaseDex:      c.BaseDexterity,
				BaseInt:      c.BaseIntelligence,
				Ascendancies: []*Ascendancy2{},
				Background: Background2{
					Active: &Size2{Width: 2000, Height: 2000}, Bg: &Size2{Width: 2000, Height: 2000},
					Image: "Classes" + c.Name, Section: "AscendancyBackground", Width: 1500, Height: 1500,
				},
			}
			for _, a := range d.ascendancies {
				if a.Character == nil || *a.Character != charIdx {
					continue
				}
				if ignoreFilter2.MatchString(a.Name) || a.Disabled {
					continue
				}
				asc := &Ascendancy2{
					ID: a.Name, Name: a.Name, InternalID: a.ID,
					Background: Background2{Image: "Classes" + a.Name, Section: "AscendancyBackground", Width: 1500, Height: 1500},
				}
				if a.BaseAscendancy != nil {
					from := d.ascendancies[*a.BaseAscendancy].Name
					if _, ok := ascendancyReplacements[from]; !ok {
						replacementOrder = append(replacementOrder, from)
					}
					ascendancyReplacements[from] = a.Name
					asc.Replace = from
				}
				class.Ascendancies = append(class.Ascendancies, asc)
			}
			if len(class.Ascendancies) == 0 {
				continue
			}
			tree.Classes = append(tree.Classes, class)
		}
	}

	// passivetree.lua:612-652
	baseOptions := make([]NodeOption2, 0, len(baseAttributes2))
	for _, id := range baseAttributes2 {
		p := row(id)
		if p == nil || ignoreFilter2.MatchString(p.Name) {
			return nil, fmt.Errorf("base attribute %d not found", id)
		}
		stats, err := d.describe(p)
		if err != nil {
			return nil, err
		}
		baseOptions = append(baseOptions, NodeOption2{ID: &id, Name: p.Name, Icon: p.Icon, Stats: stats})
	}

	switchedOption := func(idx *int) (NodeOption2, error) {
		if idx == nil {
			return NodeOption2{}, errors.New("override without a switched node")
		}
		s := &d.passiveSkills[*idx]
		stats, err := d.describe(s)
		if err != nil {
			return NodeOption2{}, err
		}
		id := s.PassiveSkillGraphID
		return NodeOption2{ID: &id, Name: s.Name, Icon: s.Icon, Stats: stats}, nil
	}

	// passivetree.lua:654-1081
	type ascendancyGroup struct {
		groups  []int
		startID uint32
	}
	ascendancyGroups := make(map[string]*ascendancyGroup)
	orbitsConstants := make(map[int]int)
	tree.Groups = make([]*Group2, len(graph.Groups))
	treeMinMax := func(x, y float64) {
		tree.MinX = min(tree.MinX, x)
		tree.MinY = min(tree.MinY, y)
		tree.MaxX = max(tree.MaxX, x)
		tree.MaxY = max(tree.MaxY, y)
	}
	for gi, group := range graph.Groups {
		i := gi + 1
		groupIsAscendancy := false
		treeGroup := &Group2{X: roundTo(float64(group.X), 2), Y: roundTo(float64(group.Y), 2), Orbits: []int{}, Nodes: []uint32{}}
		orbits := make(map[int]bool)

		for _, passive := range group.Nodes {
			node := &Node2{
				Skill:       passive.ID,
				Group:       i,
				Orbit:       passive.Orbit,
				OrbitIndex:  passive.OrbitIndex,
				Connections: []Connection2{},
			}

			if pi, ok := d.bySkillID[passive.ID]; ok {
				p := &d.passiveSkills[pi]
				if p.Name == "" {
					continue
				}
				if ignoreFilter2.MatchString(p.Name) && !slices.Contains(dntKept2, p.Name) {
					continue
				}
				name := escapeGGGString(p.Name)
				node.Name = &name
				node.StringID = p.ID
				node.Icon = &p.Icon
				if p.FlavourText != "" {
					node.FlavourText = strings.ReplaceAll(strings.ReplaceAll(p.FlavourText, "\r", ""), "\n", `\n`)
				}
				switch {
				case p.IsKeystone:
					node.IsKeystone = true
				case p.IsNotable:
					node.IsNotable = true
				case p.IsJustIcon:
					node.IsOnlyImage = true
				case p.IsJewelSocket:
					node.IsJewelSocket = true
				}

				if p.IsJewelSocket && p.IsAnointmentOnly {
					node.AliasPassiveSocket = p.ID
					node.Sinister = true
					node.NoRadius = true
				}

				if p.Ascendancy != nil {
					groupIsAscendancy = true
					a := &d.ascendancies[*p.Ascendancy]
					if ignoreFilter2.MatchString(a.Name) || a.Disabled {
						continue
					}
					node.AscendancyName = a.Name
					node.IsAscendancyStart = p.IsAscendancyStart
					if node.IsAscendancyStart {
						n := a.Name
						node.Name = &n
					}

					if p.IsJewelSocket {
						node.ContainJewelSocket = true
						if a.UIArt != nil && d.uiArt[*a.UIArt].JewelFrame != nil {
							overlay, err := d.frameOverlay(*d.uiArt[*a.UIArt].JewelFrame)
							if err != nil {
								return nil, err
							}
							node.NodeOverlay = overlay
						}
					} else if !node.IsOnlyImage {
						node.NodeOverlay = ascendancyOverlay(a.Name, node.IsNotable)
					}

					ag := ascendancyGroups[a.Name]
					if ag == nil {
						ag = &ascendancyGroup{}
						ascendancyGroups[a.Name] = ag
					}
					if p.IsAscendancyStart {
						ag.startID = passive.ID
					}
					if !slices.Contains(ag.groups, i) {
						ag.groups = append(ag.groups, i)
					}
				}

				if p.NodeFrameArt != nil && !node.IsOnlyImage {
					overlay, err := d.frameOverlay(*p.NodeFrameArt)
					if err != nil {
						return nil, err
					}
					node.NodeOverlay = overlay
				}

				if len(p.UnlockedBy) > 0 {
					node.UnlockConstraint = &UnlockConstraint2{Nodes: []uint32{}}
					if p.VisibleForAscendancy != nil {
						node.UnlockConstraint.Ascendancy = d.ascendancies[*p.VisibleForAscendancy].Name
					}
					for _, ref := range p.UnlockedBy {
						node.UnlockConstraint.Nodes = append(node.UnlockConstraint.Nodes, d.passiveSkills[ref].PassiveSkillGraphID)
					}
					node.ConnectionArt = "CharacterPlanned"
				}

				stats, err := d.describe(p)
				if err != nil {
					return nil, err
				}
				node.Stats = append([]string{}, stats...)

				if p.MasteryGroup != nil {
					if art := d.masteryGroups[*p.MasteryGroup].Art; art != nil {
						node.ActiveEffectImage = d.masteryArt[*art].ActiveEffectImage
					}
				}

				if p.IsAttribute {
					node.Options = baseOptions
					node.IsAttribute = true
				}

				if p.GrantedSkill != nil {
					gem := d.skillGems[*p.GrantedSkill]
					if gem.BaseItemType == nil {
						return nil, fmt.Errorf("granted skill of %s has no base item", p.ID)
					}
					for range gem.GemEffects {
						node.Stats = append(node.Stats, "Grants Skill: "+d.baseItemTypes[*gem.BaseItemType].Name)
					}
				}

				if p.SkillPointsGranted > 0 {
					node.Stats = append(node.Stats, "Grants "+strconv.Itoa(p.SkillPointsGranted)+" Passive Skill Point")
				}

				if p.WeaponPointsGranted > 0 {
					node.Stats = append(node.Stats, strconv.Itoa(p.WeaponPointsGranted)+" Passive Skill Points become Weapon Set Skill Points")
				}

				if ri := slices.IndexFunc(d.blightResults, func(r blightResult2) bool { return r.PassiveSkill != nil && *r.PassiveSkill == pi }); ri >= 0 {
					node.Recipe = []string{}
					if ci := slices.IndexFunc(d.blightRecipes, func(r blightRecipe2) bool { return r.Result != nil && *r.Result == ri }); ci >= 0 {
						for _, item := range d.blightRecipes[ci].Items {
							node.Recipe = append(node.Recipe, d.blightItems[item].NameShort)
						}
					}
				}

				if oi := slices.IndexFunc(d.classOverrides, func(o classOverride2) bool { return o.Original != nil && *o.Original == pi }); oi >= 0 {
					o := d.classOverrides[oi]
					node.IsSwitchable = true
					info, err := switchedOption(o.Switched)
					if err != nil {
						return nil, err
					}
					if o.Character == nil {
						return nil, fmt.Errorf("class override of %s has no character", p.ID)
					}
					node.Options = map[string]NodeOption2{d.characters[*o.Character].Name: info}
				}

				if p.Ascendancy != nil {
					if replacement, ok := ascendancyReplacements[d.ascendancies[*p.Ascendancy].Name]; ok {
						node.IsSwitchable = true
						ai := slices.IndexFunc(d.ascendancies, func(a ascendancy2) bool { return a.Name == replacement })
						info := NodeOption2{AscendancyName: replacement}
						if oi := slices.IndexFunc(d.ascendancyOverrides, func(o ascendancyOverride2) bool { return o.Original != nil && *o.Original == pi }); oi >= 0 {
							var err error
							info, err = switchedOption(d.ascendancyOverrides[oi].Switched)
							if err != nil {
								return nil, err
							}
							info.AscendancyName = replacement
						}
						a := d.ascendancies[ai]
						if p.IsJewelSocket && a.UIArt != nil && d.uiArt[*a.UIArt].JewelFrame != nil {
							overlay, err := d.frameOverlay(*d.uiArt[*a.UIArt].JewelFrame)
							if err != nil {
								return nil, err
							}
							info.NodeOverlay = overlay
						} else if !node.IsOnlyImage {
							// passivetree.lua:1008 reads nodeInfo.isNotable, which is never set.
							info.NodeOverlay = ascendancyOverlay(a.Name, false)
						}
						node.Options = map[string]NodeOption2{replacement: info}
					}
				}

				if len(p.Characters) > 0 {
					node.ClassesStart = []string{}
					for _, c := range p.Characters {
						if !ignoreFilter2.MatchString(d.characters[c].Name) {
							node.ClassesStart = append(node.ClassesStart, d.characters[c].Name)
						}
					}
				}

				node.IsMultipleChoice = p.IsMultipleChoice
				node.IsMultipleChoiceOption = p.IsMultipleChoiceOption
				node.IsFreeAllocate = p.IsFree
				node.ApplyToArmour = p.ApplyToArmour
			}

			for k, out := range passive.Out {
				node.Connections = append(node.Connections, Connection2{ID: out, Orbit: passive.OutOrbit[k]})
			}

			r := int(passive.Orbit)
			orbits[r] = true
			orbitsConstants[r] = max(orbitsConstants[r], 1, int(passive.OrbitIndex))
			tree.Nodes[strconv.FormatUint(uint64(passive.ID), 10)] = node
			treeGroup.Nodes = append(treeGroup.Nodes, passive.ID)
		}

		for orbit := range orbits {
			treeGroup.Orbits = append(treeGroup.Orbits, orbit)
		}
		slices.Sort(treeGroup.Orbits)

		if len(treeGroup.Nodes) > 0 {
			tree.Groups[gi] = treeGroup
			// passivetree.lua:1073-1076 reads the psg's unrounded position.
			if !groupIsAscendancy {
				treeMinMax(float64(group.X), float64(group.Y))
			}
		}
	}

	// passivetree.lua:1093-1098
	tree.JewelSlots = []uint32{}
	for _, slot := range d.jewelSlots {
		if slot.Slot == nil {
			return nil, errors.New("jewel slot without a passive skill")
		}
		tree.JewelSlots = append(tree.JewelSlots, d.passiveSkills[*slot.Slot].PassiveSkillGraphID)
	}

	// passivetree.lua:1100-1111; ipairs stops at the first orbit no node uses.
	for i := 0; ; i++ {
		n, ok := orbitsConstants[i]
		if !ok {
			break
		}
		if i > 0 {
			n = (n + 11) / 12 * 12
		}
		tree.Constants.SkillsPerOrbit = append(tree.Constants.SkillsPerOrbit, n)
		tree.Constants.OrbitAnglesByOrbit = append(tree.Constants.OrbitAnglesByOrbit, calcOrbitAngles(n))
	}

	// passivetree.lua:1113-1197
	radiusTree := (max(tree.MaxX-tree.MinX, tree.MaxY-tree.MinY) + 700) / 2
	hardCoded := radiusTree + 2800
	for _, classID := range graph.Roots {
		nodeStart := tree.Nodes[strconv.FormatUint(uint64(classID), 10)]
		if nodeStart == nil || tree.Groups[nodeStart.Group-1] == nil {
			return nil, fmt.Errorf("class start %d is not on the tree", classID)
		}
		group := tree.Groups[nodeStart.Group-1]
		angleToCenter := math.Atan2(group.Y, group.X)

		total := 0
		var classes []*Class2
		for _, class := range tree.Classes {
			for _, name := range nodeStart.ClassesStart {
				if name == class.Name {
					total += len(class.Ascendancies)
					classes = append(classes, class)
				}
			}
		}

		startAngle := angleToCenter - arcAngle2[total]/2*math.Pi/180
		angleStep := arcAngle2[total] / float64(total-1) * math.Pi / 180
		j := 1
		for _, class := range classes {
			for _, asc := range class.Ascendancies {
				angle := startAngle + float64(j-1)*angleStep
				cX := hardCoded * math.Cos(angle)
				cY := hardCoded * math.Sin(angle)

				info := ascendancyGroups[asc.ID]
				if info == nil {
					continue
				}
				ascendancyNode := tree.Nodes[strconv.FormatUint(uint64(info.startID), 10)]
				if ascendancyNode == nil {
					continue
				}

				asc.Background.X = cX
				asc.Background.Y = cY

				groupAscendancy := tree.Groups[ascendancyNode.Group-1]
				ai := slices.IndexFunc(d.ascendancies, func(a ascendancy2) bool { return a.ID == asc.InternalID })
				innerRadius := float64(d.ascendancies[ai].TreeRegionVector)

				newInnerX := cX + math.Cos(angleToCenter)*innerRadius
				newInnerY := cY + math.Sin(angleToCenter)*innerRadius

				nodeAngle := tree.Constants.OrbitAnglesByOrbit[ascendancyNode.Orbit][ascendancyNode.OrbitIndex]
				orbitRadius := tree.Constants.OrbitRadii[ascendancyNode.Orbit]
				newX := newInnerX - math.Sin(nodeAngle)*orbitRadius
				newY := newInnerY + math.Cos(nodeAngle)*orbitRadius

				offsetX := newX - groupAscendancy.X
				offsetY := newY - groupAscendancy.Y

				for _, gid := range info.groups {
					g := tree.Groups[gid-1]
					if g == nil {
						continue
					}
					g.X += offsetX
					g.Y += offsetY
					tree.MinX = min(tree.MinX, g.X-hardCoded/2)
					tree.MinY = min(tree.MinY, g.Y-hardCoded/2)
					tree.MaxX = max(tree.MaxX, g.X+hardCoded/2)
					tree.MaxY = max(tree.MaxY, g.Y+hardCoded/2)
				}

				j++
			}
		}
	}

	// passivetree.lua:1199-1229
	for _, from := range replacementOrder {
		to := ascendancyReplacements[from]
		var fromAsc, toAsc *Ascendancy2
		for _, class := range tree.Classes {
			for _, asc := range class.Ascendancies {
				if asc.Name == from {
					fromAsc = asc
				}
				if asc.Name == to {
					toAsc = asc
				}
			}
		}
		if fromAsc == nil || toAsc == nil {
			continue
		}
		fromAsc.ReplaceBy = to
		toAsc.Background.X = fromAsc.Background.X
		toAsc.Background.Y = fromAsc.Background.Y
	}

	return tree, nil
}

// Marshal encodes the tree as data.json.
func (t *Tree2) Marshal() ([]byte, error) {
	b, err := json.Marshal(t)
	return b, errors.Wrap(err, "failed to marshal tree")
}

// WritePoE2 generates the PoE2 tree and writes it to dir/data.json.
func WritePoE2(loader fs.FS, gameVersion string, dir string) (*Tree2, error) {
	tree, err := GeneratePoE2(loader, gameVersion)
	if err != nil {
		return nil, err
	}
	b, err := tree.Marshal()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, errors.Wrap(err, "failed to create tree directory")
	}
	return tree, errors.Wrap(os.WriteFile(filepath.Join(dir, "data.json"), b, 0o644), "failed to write data.json")
}
