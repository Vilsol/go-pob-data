package treegen

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
)

// Field order in these structs is the key order of GGG's skilltree-export data.json.

//nolint:govet // fields follow the export key order
type Tree struct {
	Tree                  string        `json:"tree"`
	Classes               []*Class      `json:"classes"`
	AlternateAscendancies []*Ascendancy `json:"alternate_ascendancies,omitempty"`
	Groups                Groups        `json:"groups"`
	Nodes                 Nodes         `json:"nodes"`
	JewelSlots            []uint32      `json:"jewelSlots"`
	MinX                  int           `json:"min_x"`
	MinY                  int           `json:"min_y"`
	MaxX                  int           `json:"max_x"`
	MaxY                  int           `json:"max_y"`
	Constants             Constants     `json:"constants"`
	Points                Points        `json:"points"`
}

//nolint:govet // fields follow the export key order
type Class struct {
	Name         string        `json:"name"`
	BaseStr      int           `json:"base_str"`
	BaseDex      int           `json:"base_dex"`
	BaseInt      int           `json:"base_int"`
	Ascendancies []*Ascendancy `json:"ascendancies"`
}

//nolint:govet // fields follow the export key order
type Ascendancy struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	FlavourText       string `json:"flavourText,omitempty"`
	FlavourTextColour string `json:"flavourTextColour,omitempty"`
	FlavourTextRect   *Rect  `json:"flavourTextRect,omitempty"`
}

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

//nolint:govet // fields follow the export key order
type Group struct {
	X          Float            `json:"x"`
	Y          Float            `json:"y"`
	IsProxy    bool             `json:"isProxy,omitempty"`
	Orbits     []int            `json:"orbits"`
	Background *GroupBackground `json:"background,omitempty"`
	Nodes      []string         `json:"nodes"`
}

//nolint:govet // fields follow the export key order
type GroupBackground struct {
	Image       string `json:"image"`
	IsHalfImage bool   `json:"isHalfImage,omitempty"`
}

//nolint:govet // fields follow the export key order
type Node struct {
	Skill                  uint32           `json:"skill"`
	Name                   string           `json:"name"`
	Icon                   string           `json:"icon,omitempty"`
	IsKeystone             bool             `json:"isKeystone,omitempty"`
	IsNotable              bool             `json:"isNotable,omitempty"`
	IsMastery              bool             `json:"isMastery,omitempty"`
	InactiveIcon           string           `json:"inactiveIcon,omitempty"`
	ActiveIcon             string           `json:"activeIcon,omitempty"`
	ActiveEffectImage      string           `json:"activeEffectImage,omitempty"`
	MasteryEffects         []*MasteryEffect `json:"masteryEffects,omitempty"`
	IsJewelSocket          bool             `json:"isJewelSocket,omitempty"`
	ExpansionJewel         *ExpansionJewel  `json:"expansionJewel,omitempty"`
	IsBlighted             bool             `json:"isBlighted,omitempty"`
	Recipe                 []string         `json:"recipe,omitempty"`
	GrantedStrength        int              `json:"grantedStrength,omitempty"`
	GrantedDexterity       int              `json:"grantedDexterity,omitempty"`
	GrantedIntelligence    int              `json:"grantedIntelligence,omitempty"`
	AscendancyName         string           `json:"ascendancyName,omitempty"`
	IsAscendancyStart      bool             `json:"isAscendancyStart,omitempty"`
	IsBloodline            bool             `json:"isBloodline,omitempty"`
	IsMultipleChoice       bool             `json:"isMultipleChoice,omitempty"`
	IsMultipleChoiceOption bool             `json:"isMultipleChoiceOption,omitempty"`
	GrantedPassivePoints   int              `json:"grantedPassivePoints,omitempty"`
	IsProxy                bool             `json:"isProxy,omitempty"`
	Stats                  []string         `json:"stats"`
	ReminderText           []string         `json:"reminderText,omitempty"`
	FlavourText            []string         `json:"flavourText,omitempty"`
	ClassStartIndex        *int             `json:"classStartIndex,omitempty"`
	// Undescribed are the node's stats without a description, which the game doesn't show.
	Undescribed []string `json:"-"`
	// Position is nil for nodes outside the graph, which omits all of its keys.
	*Position
}

//nolint:govet // fields follow the export key order
type Position struct {
	Group      int      `json:"group"`
	Orbit      uint32   `json:"orbit"`
	OrbitIndex uint32   `json:"orbitIndex"`
	Out        []string `json:"out"`
	In         []string `json:"in"`
}

//nolint:govet // fields follow the export key order
type RootNode struct {
	Group      int      `json:"group"`
	Orbit      int      `json:"orbit"`
	OrbitIndex int      `json:"orbitIndex"`
	Out        []string `json:"out"`
	In         []string `json:"in"`
}

//nolint:govet // fields follow the export key order
type MasteryEffect struct {
	Effect       uint16   `json:"effect"`
	Stats        []string `json:"stats"`
	ReminderText []string `json:"reminderText,omitempty"`
}

//nolint:govet // fields follow the export key order
type ExpansionJewel struct {
	Size   int    `json:"size"`
	Index  int    `json:"index"`
	Proxy  string `json:"proxy"`
	Parent string `json:"parent,omitempty"`
}

//nolint:govet // fields follow the export key order
type Constants struct {
	Classes              Ordered[int] `json:"classes"`
	CharacterAttributes  Ordered[int] `json:"characterAttributes"`
	PSSCentreInnerRadius int          `json:"PSSCentreInnerRadius"`
	SkillsPerOrbit       []int        `json:"skillsPerOrbit"`
	OrbitRadii           []int        `json:"orbitRadii"`
}

type Points struct {
	TotalPoints      int `json:"totalPoints"`
	AscendancyPoints int `json:"ascendancyPoints"`
}

// Float marshals like the export: 6 significant digits, never in exponent form.
type Float float32

func (f Float) MarshalJSON() ([]byte, error) {
	rounded, _ := strconv.ParseFloat(strconv.FormatFloat(float64(f), 'g', 6, 32), 64)
	return []byte(strconv.FormatFloat(rounded, 'f', -1, 64)), nil
}

// Ordered is a JSON object that keeps insertion order.
type Ordered[V any] []Entry[V]

type Entry[V any] struct {
	Value V
	Key   string
}

func (o Ordered[V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, e := range o {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeJSON(&buf, e.Key); err != nil {
			return nil, err
		}
		buf.WriteByte(':')
		if err := writeJSON(&buf, e.Value); err != nil {
			return nil, err
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// Groups are keyed by their 1-based position in the graph.
type Groups []*Group

func (g Groups) MarshalJSON() ([]byte, error) {
	o := make(Ordered[*Group], len(g))
	for i, group := range g {
		o[i] = Entry[*Group]{Key: strconv.Itoa(i + 1), Value: group}
	}
	return o.MarshalJSON()
}

// Nodes are keyed by skill id, with the synthetic root first.
type Nodes struct {
	Root *RootNode
	List []*Node
}

func (n Nodes) MarshalJSON() ([]byte, error) {
	o := make(Ordered[any], 0, len(n.List)+1)
	o = append(o, Entry[any]{Key: "root", Value: n.Root})
	for _, node := range n.List {
		o = append(o, Entry[any]{Key: strconv.FormatUint(uint64(node.Skill), 10), Value: node})
	}
	return o.MarshalJSON()
}

// writeJSON encodes compactly without HTML escaping, which the export doesn't use.
func writeJSON(buf *bytes.Buffer, v any) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return errors.Wrap(err, "failed to encode json")
	}
	buf.Truncate(buf.Len() - 1)
	return nil
}

// Marshal encodes the tree in the export's layout (4-space indent, no trailing newline).
func (t *Tree) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	if err := enc.Encode(t); err != nil {
		return nil, errors.Wrap(err, "failed to encode tree")
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}
