// Package derived holds tables Path of Building's export scripts (src/Export) derive from the PoE1 game
// data alone, ported to Go and run over a version's committed tables and stat descriptions. cmd/derive
// publishes them as tuple assets under data/<version>/derived/, named by their layout and contents and
// listed in the version's manifest; Load reads them.
package derived

// Every row is encoded as a strict tuple, as raw's are.

//msgp:tuple ModScalability ModScalabilityEntry ItemModStatOrder Pantheon PantheonSoul ClusterJewels ClusterJewel ClusterJewelSkill ClusterNotableOrder ClusterOrbitOffsets LegionPassive LegionRoll TattooPassive

//msgp:ignore Asset

// The tables of this package by asset name.
const (
	ModScalabilityTable   = "ModScalability"
	ItemModStatOrderTable = "ItemModStatOrder"
	PantheonsTable        = "Pantheons"
	ClusterJewelsTable    = "ClusterJewels"
	LegionPassivesTable   = "LegionPassives"
	TattooPassivesTable   = "TattooPassives"
)

// ModScalability is one data.modScalability entry (src/Data/ModScalability.lua), sorted by line: the mod
// line with every value, and the sign before it, replaced by `#`. Entry i governs the line's i-th `#`.
type ModScalability struct {
	Line    string
	Entries []ModScalabilityEntry
}

type ModScalabilityEntry struct {
	// Formats are the stat description's format handlers of the value, nil when it has none.
	Formats    []string
	IsScalable bool
}

func (ModScalability) Table() string { return ModScalabilityTable }

// ItemModStatOrder is one key of Item.lua's uniqueModStatOrder, the lowest statOrder of a line among
// data.itemMods.ItemExclusive, sorted by line. Exact is keyed by the line lowercased with its newlines
// as spaces, Normalised by normaliseModLine (Item.lua); a line under only one of them has 0 in the other.
type ItemModStatOrder struct {
	Line       string
	Exact      float64
	Normalised float64
}

func (ItemModStatOrder) Table() string { return ItemModStatOrderTable }

// Pantheon is one data.pantheons god (src/Data/Pantheons.lua), sorted by ID. Souls are in the game's
// order: the first is the god's own soul.
type Pantheon struct {
	ID         string
	Souls      []PantheonSoul
	IsMajorGod bool
}

type PantheonSoul struct {
	Name string
	// Mods are each stat's description, its lines joined by spaces.
	Mods []string
}

func (Pantheon) Table() string { return PantheonsTable }

// ClusterJewels is data.clusterJewels (src/Data/ClusterJewels.lua), the asset's only row.
type ClusterJewels struct {
	// Jewels are sorted by base name.
	Jewels []ClusterJewel
	// Keystones are in PassiveTreeExpansionSpecialSkills order.
	Keystones []string
	// NotableSortOrder is sorted by name.
	NotableSortOrder []ClusterNotableOrder
	// OrbitOffsets are sorted by proxy node hash.
	OrbitOffsets []ClusterOrbitOffsets
}

type ClusterJewel struct {
	BaseName        string
	Size            string
	SmallIndicies   []int
	NotableIndicies []int
	SocketIndicies  []int
	// Skills are sorted by ID.
	Skills        []ClusterJewelSkill
	SizeIndex     int
	MinNodes      int
	MaxNodes      int
	TotalIndicies int
}

type ClusterJewelSkill struct {
	// ID is the passive's Id, the key of the jewel's skills table.
	ID          string
	Name        string
	Icon        string
	MasteryIcon string
	Tag         string
	Stats       []string
	Enchant     []string
}

type ClusterNotableOrder struct {
	Name  string
	Order int
}

type ClusterOrbitOffsets struct {
	// Offsets are the orbit start indices by orbit, at most three.
	Offsets   []int
	ProxyHash int64
}

func (ClusterJewels) Table() string { return ClusterJewelsTable }

// LegionPassive is one node or addition of src/Data/TimelessJewelData/LegionPassives.lua, sorted by
// Addition, then ID.
type LegionPassive struct {
	ID    string
	Dn    string
	Sd    []string
	Rolls []LegionRoll
	// Keystone is a node of the keystones table.
	Keystone bool
	// Addition is a row of the additions table, otherwise of the nodes table.
	Addition bool
}

type LegionRoll struct {
	Key   string
	Fmt   string
	Index int
	Min   float64
	Max   float64
}

func (LegionPassive) Table() string { return LegionPassivesTable }

// TattooPassive is one node of src/Data/TimelessJewelData/TattooPassives.lua, sorted by Dn, its key.
type TattooPassive struct {
	// Legacy is set for a tattoo sold at the currency exchange: whether it is out of the challenge league.
	Legacy            *bool
	ID                string
	Dn                string
	OverrideType      string
	TargetType        string
	TargetValue       string
	Icon              string
	ActiveEffectImage string
	Sd                []string
	ReminderText      []string
	MinimumConnected  int
	MaximumConnected  int
	IsKeystone        bool
}

func (TattooPassive) Table() string { return TattooPassivesTable }
