// Package pobdata holds data tables of Path of Building (PoE2) as its Data.lua loads them over this
// repository's PoE2 tables: go-pob emits them from the pinned PoB-PoE2 and they are published as
// tuple assets under data/<version>/pob/, named by the hash of their contents.
package pobdata

// Every row is encoded as a strict tuple, as raw's are.

//msgp:tuple ItemBase BaseReq BaseArmour BaseWeapon BaseFlask BaseCharm ItemMod ExclusiveMod ModScalability ModScalabilityEntry

// The tables of this package by asset name.
const (
	ItemBasesTable      = "ItemBases"
	ItemModsTable       = "ItemMods"
	ExclusiveModsTable  = "ExclusiveMods"
	ModScalabilityTable = "ModScalability"
)

// ItemBase is one data.itemBases entry. A numeric field PoB leaves nil reads 0 here; go-pob's emitter
// refuses a present 0 on every field PoB tests for presence, so 0 always means absent on those.
type ItemBase struct {
	Armour           *BaseArmour
	Weapon           *BaseWeapon
	Flask            *BaseFlask
	Charm            *BaseCharm
	Name             string
	Type             string
	SubType          string
	Implicit         string
	ImplicitModTypes [][]string
	Tags             []string
	VariantList      []string
	Req              BaseReq
	SocketLimit      float64
	Quality          float64
	Spirit           float64
	CharmLimit       float64
	Hidden           bool
	// HasCharmLimit is a present charmLimit, which belts carry as 0.
	HasCharmLimit                  bool
	GrantedSkillsHaveNoReservation bool
}

type BaseReq struct {
	Level, Str, Dex, Int float64
}

type BaseArmour struct {
	Armour, Evasion, EnergyShield, Ward, MovementPenalty, BlockChance float64
}

type BaseWeapon struct {
	PhysicalMin, PhysicalMax   float64
	FireMin, FireMax           float64
	ColdMin, ColdMax           float64
	LightningMin, LightningMax float64
	ChaosMin, ChaosMax         float64
	CritChanceBase             float64
	AttackRateBase             float64
	Range                      float64
	ReloadTimeBase             float64
}

type BaseFlask struct {
	Life, Mana, Duration, ChargesMax, ChargesUsed float64
}

type BaseCharm struct {
	Buff                              []string
	Duration, ChargesMax, ChargesUsed float64
}

// ItemMod is one mod of a data.itemMods affix table (src/Data/Mod*.lua), the tables flattened to one
// row per mod sorted by category and id.
type ItemMod struct {
	Category            string
	ID                  string
	Type                string
	Affix               string
	Group               string
	Lines               []string
	StatOrder           []float64
	WeightKey           []string
	WeightVal           []float64
	WeightMultiplierKey []string
	WeightMultiplierVal []float64
	Tags                []string
	ModTags             []string
	Level               float64
	NodeType            float64
	Unscalable          bool
}

// ExclusiveMod is one data.itemMods.Exclusive entry's lines and stat orders, sorted by id.
type ExclusiveMod struct {
	ID        string
	Lines     []string
	StatOrder []float64
}

// ModScalability is one data.modScalability entry, sorted by line: entry i governs the line's i-th `#`.
type ModScalability struct {
	Line    string
	Entries []ModScalabilityEntry
}

type ModScalabilityEntry struct {
	Formats    []string
	IsScalable bool
}
