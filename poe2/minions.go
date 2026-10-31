package poe2

import (
	"context"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/Vilsol/go-pob-data/loader"
)

const (
	MonsterVarietiesTable   = "MonsterVarieties"
	MonsterTypesTable       = "MonsterTypes"
	MonsterResistancesTable = "MonsterResistances"
	MonsterCategoriesTable  = "MonsterCategories"
	MonsterPacksTable       = "MonsterPacks"
	MonsterPackEntriesTable = "MonsterPackEntries"
	WorldAreasTable         = "WorldAreas"
	EndgameMapsTable        = "EndgameMaps"
	ModsTable               = "Mods"
)

type MonsterVariety struct {
	MainHandItemClass    *int    `json:"MainHand_ItemClass"` // MainHandItemClass
	OffHandItemClass     *int    `json:"OffHand_ItemClass"`  // OffHandItemClass
	MonsterCategory      *int    `json:"MonsterCategory"`
	ID                   string  `json:"Id"`
	Name                 string  `json:"Name"`
	ObjectType           string  `json:"BaseMonsterTypeIndex"` // ObjectType
	Mods                 []int   `json:"Mods"`
	SpecialMods          []int   `json:"Special_Mods"` // SpecialMods
	Tags                 []int   `json:"Tags"`
	GrantedEffects       []int   `json:"GrantedEffects"`
	Key                  int     `json:"_key"`
	MonsterType          int     `json:"MonsterType"` // Type
	MovementSpeed        int     `json:"MovementSpeed"`
	MaximumAttackRange   int     `json:"MaximumAttackDistance"` // MaximumAttackRange
	ExperienceMultiplier int     `json:"ExperienceMultiplier"`
	DamageMultiplier     int     `json:"DamageMultiplier"`
	LifeMultiplier       int     `json:"LifeMultiplier"`
	AttackDuration       int     `json:"AttackSpeed"` // AttackDuration
	AttackCrit           float32 `json:"AttackCrit"`
}

type MonsterType struct {
	ID                           string `json:"Id"`
	Resistances                  []int  `json:"MonsterResistances"` // Resistances
	Key                          int    `json:"_key"`
	Armour                       int    `json:"Armour"`
	Evasion                      int    `json:"Evasion"`
	EnergyShield                 int    `json:"EnergyShieldFromLife"` // EnergyShield
	DamageSpread                 int    `json:"DamageSpread"`
	IsPlayerMinion               bool   `json:"IsSummoned"` // IsPlayerMinion
	BaseDamageIgnoresAttackSpeed bool   `json:"Var8"`       // BaseDamageIgnoresAttackSpeed
}

type MonsterResistance struct {
	ID              string `json:"Id"`
	NormalFire      []int  `json:"Fire1"`      // NormalFire
	RareFire        []int  `json:"Fire3"`      // RareFire
	NormalCold      []int  `json:"Cold1"`      // NormalCold
	RareCold        []int  `json:"Cold3"`      // RareCold
	NormalLightning []int  `json:"Lightning1"` // NormalLightning
	RareLightning   []int  `json:"Lightning3"` // RareLightning
	NormalChaos     []int  `json:"Chaos1"`     // NormalChaos
	RareChaos       []int  `json:"Chaos3"`     // RareChaos
	Key             int    `json:"_key"`
}

type MonsterCategory struct {
	Type string `json:"Name"` // Type
	Key  int    `json:"_key"`
	Tag  int    `json:"Tag"`
}

type MonsterPack struct {
	ID                 string `json:"Id"`
	WorldAreas         []int  `json:"WorldAreas"`
	BossMonsters       []int  `json:"BossMonsters"`
	AdditionalMonsters []int  `json:"AdditionalMonsters"`
	Key                int    `json:"_key"`
}

type MonsterPackEntry struct {
	MonsterPacksKey     *int `json:"MonsterPacksKey"`
	MonsterVarietiesKey *int `json:"MonsterVarietiesKey"`
	Key                 int  `json:"_key"`
}

type WorldArea struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Tags []int  `json:"Tags"`
	Key  int    `json:"_key"`
	Act  int    `json:"Act"`
}

type EndgameMap struct {
	NativePacks []int `json:"MonsterPacks"` // NativePacks
	Key         int   `json:"_key"`
	WorldArea   int   `json:"WorldArea"` // Id
}

// Mod declares the first six stats, the ones minions.lua reads; Stat<n>Value is the low end of the range.
type Mod struct {
	Stat1      *int   `json:"Stat1"`
	Stat2      *int   `json:"Stat2"`
	Stat3      *int   `json:"Stat3"`
	Stat4      *int   `json:"Stat4"`
	Stat5      *int   `json:"Stat5"`
	Stat6      *int   `json:"Stat6"`
	ID         string `json:"Id"`
	Key        int    `json:"_key"`
	Stat1Value int    `json:"Stat1ValueMin"`
	Stat2Value int    `json:"Stat2ValueMin"`
	Stat3Value int    `json:"Stat3ValueMin"`
	Stat4Value int    `json:"Stat4ValueMin"`
	Stat5Value int    `json:"Stat5ValueMin"`
	Stat6Value int    `json:"Stat6ValueMin"`
}

func (m *Mod) stats() [6]*int { return [6]*int{m.Stat1, m.Stat2, m.Stat3, m.Stat4, m.Stat5, m.Stat6} }
func (m *Mod) values() [6]int {
	return [6]int{m.Stat1Value, m.Stat2Value, m.Stat3Value, m.Stat4Value, m.Stat5Value, m.Stat6Value}
}

// MinionTables holds the tables PoB-PoE2's minion exporter reads (src/Export/Scripts/minions.lua).
type MinionTables struct {
	varietyByID map[string]int
	packByID    map[string]int

	MonsterVarieties   []MonsterVariety
	MonsterTypes       []MonsterType
	MonsterResistances []MonsterResistance
	MonsterCategories  []MonsterCategory
	MonsterPacks       []MonsterPack
	MonsterPackEntries []MonsterPackEntry
	WorldAreas         []WorldArea
	EndgameMaps        []EndgameMap
	Mods               []Mod
	Stats              []Stat
	Tags               []Tag
	ItemClasses        []ItemClass
	GrantedEffects     []GrantedEffect
}

// LoadMinionTables loads version's minion tables and indexes them.
func LoadMinionTables(ctx context.Context, version string, cache loader.AssetCache) (*MinionTables, error) {
	m := &MinionTables{}
	for _, load := range []func() error{
		func() error { return loadTable(ctx, version, MonsterVarietiesTable, cache, &m.MonsterVarieties) },
		func() error { return loadTable(ctx, version, MonsterTypesTable, cache, &m.MonsterTypes) },
		func() error { return loadTable(ctx, version, MonsterResistancesTable, cache, &m.MonsterResistances) },
		func() error { return loadTable(ctx, version, MonsterCategoriesTable, cache, &m.MonsterCategories) },
		func() error { return loadTable(ctx, version, MonsterPacksTable, cache, &m.MonsterPacks) },
		func() error { return loadTable(ctx, version, MonsterPackEntriesTable, cache, &m.MonsterPackEntries) },
		func() error { return loadTable(ctx, version, WorldAreasTable, cache, &m.WorldAreas) },
		func() error { return loadTable(ctx, version, EndgameMapsTable, cache, &m.EndgameMaps) },
		func() error { return loadTable(ctx, version, ModsTable, cache, &m.Mods) },
		func() error { return loadTable(ctx, version, StatsTable, cache, &m.Stats) },
		func() error { return loadTable(ctx, version, TagsTable, cache, &m.Tags) },
		func() error { return loadTable(ctx, version, ItemClassesTable, cache, &m.ItemClasses) },
		func() error { return loadTable(ctx, version, GrantedEffectsTable, cache, &m.GrantedEffects) },
	} {
		if err := load(); err != nil {
			return nil, err
		}
	}
	m.varietyByID = firstByID(m.MonsterVarieties, func(r *MonsterVariety) string { return r.ID })
	m.packByID = firstByID(m.MonsterPacks, func(r *MonsterPack) string { return r.ID })
	return m, nil
}

// firstByID is DatFile:GetRow("Id", id): the first row with that id.
func firstByID[T any](table []T, id func(*T) string) map[string]int {
	out := make(map[string]int, len(table))
	for i := range table {
		if _, ok := out[id(&table[i])]; !ok {
			out[id(&table[i])] = i
		}
	}
	return out
}

// VarietyByID is the MonsterVarieties row with id, or nil.
func (m *MinionTables) VarietyByID(id string) *MonsterVariety {
	index, ok := m.varietyByID[id]
	if !ok {
		return nil
	}
	return &m.MonsterVarieties[index]
}

// MinionModStat is one stat of a minion's modList as minions.lua:354-372 comments it: the mod (or, for an
// .ot stat, the stat) id, the stat id and its value.
type MinionModStat struct {
	ModID  string
	StatID string
	Value  float64
}

// Minion is the game-data half of one minions[...] entry of src/Data/Minions.lua or Spectres.lua. The
// template half stays PoB's: the entry name, limit, extraFlags, the #mod lines, the extra skills, and
// SkillStatMap's translation of ModList into mods.
type Minion struct {
	Name                         string
	WeaponType1                  string // "" when the entry has none
	WeaponType2                  string
	MonsterCategory              string // "" when the entry has none
	MonsterTags                  []string
	SpawnLocation                []string
	SkillList                    []string
	ModList                      []MinionModStat
	Life                         float64
	EnergyShield                 float64 // 0 when the entry has none
	Armour                       float64
	Evasion                      float64
	Damage                       float64
	DamageSpread                 float64
	AttackTime                   float64
	CritChance                   float64
	DamageFixup                  float64 // 0 when the entry has none
	CompanionReservation         float64
	FireResist                   int
	ColdResist                   int
	LightningResist              int
	ChaosResist                  int
	CompanionFireResist          int
	CompanionColdResist          int
	CompanionLightningResist     int
	CompanionChaosResist         int
	AttackRange                  int
	BaseMovementSpeed            int
	SpectreReservation           int
	BaseDamageIgnoresAttackSpeed bool
}

// itemClassMap is minions.lua:82-104.
var itemClassMap = map[string]string{
	"Claw":                     "Claw",
	"Dagger":                   "Dagger",
	"Wand":                     "Wand",
	"One Hand Sword":           "One Hand Sword",
	"Thrusting One Hand Sword": "One Hand Sword",
	"One Hand Axe":             "One Hand Axe",
	"One Hand Mace":            "One Hand Mace",
	"Crossbow":                 "Crossbow",
	"Bow":                      "Bow",
	"Fishing Rod":              "Fishing Rod",
	"Staff":                    "Staff",
	"Warstaff":                 "Warstaff",
	"Two Hand Sword":           "Two Hand Sword",
	"Two Hand Axe":             "Two Hand Axe",
	"Two Hand Mace":            "Two Hand Mace",
	"Shield":                   "Shield",
	"Sceptre":                  "One Hand Mace",
	"Flail":                    "Flail",
	"Spear":                    "Spear",
	"Talisman":                 "Talisman",
	"Unarmed":                  "None",
}

var (
	otExtends = regexp.MustCompile(`extends "(.+)"`)
	otStat    = regexp.MustCompile(`^(.*?)=(.+)$`)
	sanctumID = regexp.MustCompile(`^Sanctum_(\d+)`)
)

// Minion is the #emit directive of src/Export/Scripts/minions.lua:156-377 for the variety with id, less its
// template half. readOT returns the UTF-8 text of a game .ot file (loader.LoadFile), and false when there is
// none.
func (m *MinionTables) Minion(id string, readOT func(path string) (string, bool)) (*Minion, error) {
	monsterVariety := m.VarietyByID(id)
	if monsterVariety == nil {
		return nil, errors.New("invalid variety: " + id)
	}
	monsterType := &m.MonsterTypes[monsterVariety.MonsterType]
	if len(monsterType.Resistances) == 0 {
		return nil, errors.New(id + " has no resistances")
	}
	resistances := &m.MonsterResistances[monsterType.Resistances[0]]

	out := &Minion{
		Name:                         monsterVariety.Name,
		Life:                         float64(monsterVariety.LifeMultiplier) / 100,
		BaseDamageIgnoresAttackSpeed: monsterType.BaseDamageIgnoresAttackSpeed,
		EnergyShield:                 float64(monsterType.EnergyShield) / 100,
		Armour:                       float64(monsterType.Armour) / 100,
		Evasion:                      float64(monsterType.Evasion) / 100,
		FireResist:                   resistances.NormalFire[0],
		ColdResist:                   resistances.NormalCold[0],
		LightningResist:              resistances.NormalLightning[0],
		ChaosResist:                  resistances.NormalChaos[0],
		CompanionFireResist:          resistances.RareFire[0],
		CompanionColdResist:          resistances.RareCold[0],
		CompanionLightningResist:     resistances.RareLightning[0],
		CompanionChaosResist:         resistances.RareChaos[0],
		Damage:                       float64(monsterVariety.DamageMultiplier) / 100,
		DamageSpread:                 float64(monsterType.DamageSpread) / 100,
		AttackTime:                   float64(monsterVariety.AttackDuration) / 1000,
		AttackRange:                  monsterVariety.MaximumAttackRange,
		CritChance:                   round(float64(monsterVariety.AttackCrit)*100, 2),
		BaseMovementSpeed:            monsterVariety.MovementSpeed,
		SpectreReservation:           int(math.Floor(math.Pow(float64(monsterVariety.ExperienceMultiplier)/100, 0.75) * 50)),
		CompanionReservation:         math.Floor(math.Sqrt(float64(monsterVariety.ExperienceMultiplier)/100)*100) / 100 * 30,
		SpawnLocation:                m.spawnLocation(monsterVariety),
	}
	for _, tag := range monsterVariety.Tags {
		out.MonsterTags = append(out.MonsterTags, m.Tags[tag].ID)
	}
	for _, mod := range monsterVariety.Mods {
		switch m.Mods[mod].ID {
		case "MonsterSpeedAndDamageFixupSmall":
			out.DamageFixup = 0.11
		case "MonsterSpeedAndDamageFixupLarge":
			out.DamageFixup = 0.22
		case "MonsterSpeedAndDamageFixupComplete":
			out.DamageFixup = 0.33
		}
	}
	if monsterVariety.MainHandItemClass != nil {
		out.WeaponType1 = itemClassMap[m.ItemClasses[*monsterVariety.MainHandItemClass].ID]
	}
	if monsterVariety.OffHandItemClass != nil {
		out.WeaponType2 = itemClassMap[m.ItemClasses[*monsterVariety.OffHandItemClass].ID]
	}
	if monsterVariety.MonsterCategory != nil {
		out.MonsterCategory = m.MonsterCategories[*monsterVariety.MonsterCategory].Type
	}
	for _, grantedEffect := range monsterVariety.GrantedEffects {
		out.SkillList = append(out.SkillList, m.GrantedEffects[grantedEffect].ID)
	}

	for _, mod := range slices.Concat(monsterVariety.Mods, monsterVariety.SpecialMods) {
		row := &m.Mods[mod]
		values := row.values()
		for i, stat := range row.stats() {
			if stat != nil {
				out.ModList = append(out.ModList, MinionModStat{ModID: row.ID, StatID: m.Stats[*stat].ID, Value: float64(values[i])})
			}
		}
	}
	if monsterVariety.ObjectType != "" {
		otStats, err := getOTStats(monsterVariety.ObjectType, nil, map[string]bool{}, readOT)
		if err != nil {
			return nil, err
		}
		out.ModList = append(out.ModList, otStats...)
	}
	return out, nil
}

// spawnLocation is minions.lua's world area names: the areas of every pack the variety's name is an entry,
// additional monster or boss of, and of every endgame map those packs are native to, sorted
// (minions.lua:163-256 and 324-333).
func (m *MinionTables) spawnLocation(monsterVariety *MonsterVariety) []string {
	var matchingEntries []string
	allMonsterPackIDs := map[string]bool{}
	for _, entry := range m.MonsterPackEntries {
		if entry.MonsterPacksKey == nil {
			continue
		}
		packID := m.MonsterPacks[*entry.MonsterPacksKey].ID
		allMonsterPackIDs[packID] = true
		if entry.MonsterVarietiesKey != nil && m.MonsterVarieties[*entry.MonsterVarietiesKey].Name == monsterVariety.Name {
			matchingEntries = append(matchingEntries, packID)
		}
	}
	// PoB walks a hash set here; the order only reaches worldAreaNames, which is sorted and deduplicated.
	for _, packID := range sortedKeys(allMonsterPackIDs) {
		pack := &m.MonsterPacks[m.packByID[packID]]
		for _, addMon := range pack.AdditionalMonsters {
			if m.MonsterVarieties[addMon].Name == monsterVariety.Name {
				matchingEntries = append(matchingEntries, pack.ID)
			}
		}
		for _, bossMon := range pack.BossMonsters {
			if m.MonsterVarieties[bossMon].Name == monsterVariety.Name {
				matchingEntries = append(matchingEntries, pack.ID)
			}
		}
	}

	var worldAreaNames []string
	seenAreas := map[string]bool{}
	add := func(area *WorldArea, sanctum bool) {
		if area.Name == "NULL" || strings.Contains(area.Name, "DNT") {
			return
		}
		isMap := false
		for _, tag := range area.Tags {
			if m.Tags[tag].ID == "map" {
				isMap = true
			}
		}
		displayName := area.Name
		if isMap {
			displayName += " (Map)"
		} else if floor := sanctumID.FindStringSubmatch(area.ID); sanctum && floor != nil {
			displayName += " (Floor " + floor[1] + ")"
		} else if area.Act != 10 && (!sanctum || area.Name != "Trial of the Sekhemas") {
			displayName += " (Act " + strconv.Itoa(area.Act) + ")"
		}
		if !seenAreas[displayName] {
			worldAreaNames = append(worldAreaNames, displayName)
			seenAreas[displayName] = true
		}
	}
	for _, packID := range matchingEntries {
		index, ok := m.packByID[packID]
		if ok {
			for _, worldAreaRef := range m.MonsterPacks[index].WorldAreas {
				add(&m.WorldAreas[worldAreaRef], true)
			}
		}
		for _, mapRow := range m.EndgameMaps {
			for _, nativePack := range mapRow.NativePacks {
				if m.MonsterPacks[nativePack].ID == packID {
					add(&m.WorldAreas[mapRow.WorldArea], false)
				}
			}
		}
	}

	slices.Sort(worldAreaNames)
	for i, name := range worldAreaNames {
		if name == "The Ziggurat Refuge" {
			worldAreaNames[i] = "Found in Maps"
		}
	}
	return worldAreaNames
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// getOTStats reads the Stats blocks of an .ot file and the files it extends, parents first where the
// extends line comes first
// (minions.lua:41-80).
func getOTStats(otFile string, modList []MinionModStat, visited map[string]bool, readOT func(string) (string, bool)) ([]MinionModStat, error) {
	if visited[otFile] {
		return modList, nil
	}
	visited[otFile] = true

	text, ok := readOT(otFile + ".ot")
	if !ok {
		return modList, nil
	}
	inWantedBlock := false
	for _, line := range strings.FieldsFunc(text, func(r rune) bool { return r == '\r' || r == '\n' }) {
		if superClass := otExtends.FindStringSubmatch(line); superClass != nil && superClass[1] != "Metadata/Monsters/Monster" && superClass[1] != "nothing" {
			var err error
			modList, err = getOTStats(superClass[1], modList, visited, readOT)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(line, "Stats") {
			inWantedBlock = true
		} else if inWantedBlock && strings.HasPrefix(line, "}") {
			inWantedBlock = false
		} else if inWantedBlock && strings.Contains(line, "=") && !strings.Contains(line, "//") {
			if kv := otStat.FindStringSubmatch(luaStripSpace(line)); kv != nil {
				value, err := strconv.ParseFloat(kv[2], 64)
				if err != nil {
					return nil, errors.Wrap(err, otFile+": "+line)
				}
				modList = append(modList, MinionModStat{ModID: kv[1], StatID: kv[1], Value: value})
			}
		}
	}
	return modList, nil
}

// luaStripSpace is line:gsub("%s+", "").
func luaStripSpace(line string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(" \t\n\v\f\r", r) {
			return -1
		}
		return r
	}, line)
}

// round is src/Modules/Common.lua's round.
func round(val float64, dec int) float64 {
	return math.Floor(val*math.Pow(10, float64(dec))+0.5) / math.Pow(10, float64(dec))
}
