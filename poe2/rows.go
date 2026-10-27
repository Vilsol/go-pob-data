// Package poe2 holds rows of Path of Exile 2's json tables, for loader.LoadJSONTable:
//
//	rows, err := loader.LoadJSONTable[poe2.GameConstant](ctx, "0.5", poe2.GameConstantsTable, cache)
//
// Only the columns consumers read are declared. A foreign key is the row index (Key) of the table it
// names, and a nullable one is a pointer.
package poe2

const (
	GameConstantsTable              = "GameConstants"
	PlayerMinionIntrinsicStatsTable = "PlayerMinionIntrinsicStats"
	SoulCoresTable                  = "SoulCores"
	SoulCoreStatsTable              = "SoulCoreStats"
	SoulCoreTypesTable              = "SoulCoreTypes"
	SoulCoreLimitsTable             = "SoulCoreLimits"
	SoulCoreStatCategoriesTable     = "SoulCoreStatCategories"
	QuestStaticRewardsTable         = "QuestStaticRewards"
	AlternatePassiveSkillsTable     = "AlternatePassiveSkills"
	AlternatePassiveAdditionsTable  = "AlternatePassiveAdditions"
	StatsTable                      = "Stats"
	BaseItemTypesTable              = "BaseItemTypes"
	CharacterMeleeSkillsTable       = "CharacterMeleeSkills"
)

type GameConstant struct {
	ID      string `json:"Id"`
	Key     int    `json:"_key"`
	Value   int    `json:"Value"`
	Divisor int    `json:"Divisor"`
}

type PlayerMinionIntrinsicStat struct {
	Key   int `json:"_key"`
	Stat  int `json:"Stat"`
	Value int `json:"Value"`
}

type SoulCore struct {
	Limit                          *int `json:"Limit"`
	TierHigher                     *int `json:"TierHigher"`
	Key                            int  `json:"_key"`
	BaseItemType                   int  `json:"BaseItemType"`
	RequiredLevel                  int  `json:"RequiredLevel"`
	Type                           int  `json:"Type"`
	IsSocketBound                  bool `json:"IsSocketBound"`
	CanSocketInMartialArtistSlots  bool `json:"CanSocketInMartialArtistSlots"`
	CanSocketInUniqueItems         bool `json:"CanSocketInUniqueItems"`
	CanSocketInJewellery           bool `json:"CanSocketInJewellery"`
	CanSocketInCorruptedSanctified bool `json:"CanSocketInCorruptedSanctified"`
}

type SoulCoreStat struct {
	Stats             []int `json:"Stats"`
	StatsValues       []int `json:"StatsValues"`
	BondedStats       []int `json:"BondedStats"`
	BondedStatsValues []int `json:"BondedStatsValues"`
	Key               int   `json:"_key"`
	SoulCore          int   `json:"SoulCore"`
	StatCategory      int   `json:"StatCategory"`
}

// SoulCoreType resolves SoulCore.Type.
type SoulCoreType struct {
	EffectStat   *int   `json:"EffectStat"`
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Key          int    `json:"_key"`
	SocketedStat int    `json:"SocketedStat"` // CountStat
}

// SoulCoreLimit resolves SoulCore.Limit.
type SoulCoreLimit struct {
	ID    string `json:"Id"`
	Text  string `json:"Text"` // Description
	Key   int    `json:"_key"`
	Limit int    `json:"Limit"`
}

// SoulCoreStatCategory resolves SoulCoreStat.StatCategory.
type SoulCoreStatCategory struct {
	ID                string `json:"Id"`
	Display           string `json:"Display"`           // Text
	TargetItemClasses []int  `json:"TargetItemClasses"` // ItemClass
	Key               int    `json:"_key"`
}

type QuestStaticReward struct {
	QuestKey          *int  `json:"QuestKey"`
	RewardStats       []int `json:"RewardStats"`
	RewardStatsValues []int `json:"RewardStatsValues"`
	Key               int   `json:"_key"`
	QuestFlag         int   `json:"QuestFlag"`
}

type AlternatePassiveSkill struct {
	ID                   string `json:"Id"`
	Name                 string `json:"Name"`
	FlavourText          string `json:"FlavourText"`
	DDSIcon              string `json:"DDSIcon"`
	PassiveType          []int  `json:"PassiveType"`
	Stats                []int  `json:"Stats"`
	Key                  int    `json:"_key"`
	AlternateTreeVersion int    `json:"AlternateTreeVersion"`
	Stat1Min             int    `json:"Stat1Min"`
	Stat1Max             int    `json:"Stat1Max"`
	Stat2Min             int    `json:"Stat2Min"`
	Stat2Max             int    `json:"Stat2Max"`
	Stat3Min             int    `json:"Stat3Min"`
	Stat3Max             int    `json:"Stat3Max"`
	Stat4Min             int    `json:"Stat4Min"`
	Stat4Max             int    `json:"Stat4Max"`
	Stat5Min             int    `json:"Stat5Min"`
	Stat5Max             int    `json:"Stat5Max"`
	Stat6Min             int    `json:"Stat6Min"`
	Stat6Max             int    `json:"Stat6Max"`
	SpawnWeight          int    `json:"SpawnWeight"`
	ConquerorIndex       int    `json:"ConquerorIndex"`
	RandomMin            int    `json:"RandomMin"`
	RandomMax            int    `json:"RandomMax"`
	ConquerorVersion     int    `json:"ConquerorVersion"`
	ConquerorSpawnWeight int    `json:"ConquerorSpawnWeight"`
}

type AlternatePassiveAddition struct {
	ID                   string `json:"Id"`
	Stats                []int  `json:"Stats"`
	PassiveType          []int  `json:"PassiveType"`
	Key                  int    `json:"_key"`
	AlternateTreeVersion int    `json:"AlternateTreeVersion"`
	SpawnWeight          int    `json:"SpawnWeight"`
	Stat1Min             int    `json:"Stat1Min"`
	Stat1Max             int    `json:"Stat1Max"`
	Stat2Min             int    `json:"Stat2Min"`
	Stat2Max             int    `json:"Stat2Max"`
	Stat3Min             int    `json:"Stat3Min"`
	Stat3Max             int    `json:"Stat3Max"`
}

// Stat resolves the stat keys of the other rows.
type Stat struct {
	ID                  string `json:"Id"`
	Key                 int    `json:"_key"`
	CannotGrantToMinion bool   `json:"Var13"`
}

// BaseItemType resolves SoulCore.BaseItemType.
type BaseItemType struct {
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	InheritsFrom string `json:"InheritsFrom"`  // BaseType
	ImplicitMods []int  `json:"Implicit_Mods"` // ImplicitMods
	Tags         []int  `json:"Tags"`
	Key          int    `json:"_key"`
	ItemClass    int    `json:"ItemClass"`
	DropLevel    int    `json:"DropLevel"`
}

// CharacterMeleeSkill is a default attack: the skill gems granted for a main-hand and off-hand
// WieldableClasses pair.
type CharacterMeleeSkill struct {
	SkillGems []int `json:"SkillGems"` // SkillGem
	Key       int   `json:"_key"`
	MainHand  int   `json:"MainHand"` // MainHandItem
	OffHand   int   `json:"OffHand"`  // OffHandItem
}
