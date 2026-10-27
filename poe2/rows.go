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
	QuestStaticRewardsTable         = "QuestStaticRewards"
	AlternatePassiveSkillsTable     = "AlternatePassiveSkills"
	AlternatePassiveAdditionsTable  = "AlternatePassiveAdditions"
	StatsTable                      = "Stats"
	BaseItemTypesTable              = "BaseItemTypes"
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
	ID  string `json:"Id"`
	Key int    `json:"_key"`
}

// BaseItemType resolves SoulCore.BaseItemType.
type BaseItemType struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Key  int    `json:"_key"`
}
