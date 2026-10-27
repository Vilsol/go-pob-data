package poe2

// The tables PoB-PoE2's skill exporter (src/Export/Scripts/skills.lua) reads. Columns keep the names the
// json carries; where PoB's spec names a column differently, the PoB name follows in a comment.
const (
	ActiveSkillTypeTable               = "ActiveSkillType"
	ActiveSkillsTable                  = "ActiveSkills"
	ActiveSkillWeaponRequirementTable  = "ActiveSkillWeaponRequirement"
	WieldableClassesTable              = "WieldableClasses"
	ItemClassesTable                   = "ItemClasses"
	ItemClassCategoriesTable           = "ItemClassCategories"
	GrantedEffectsTable                = "GrantedEffects"
	GrantedEffectsPerLevelTable        = "GrantedEffectsPerLevel"
	GrantedEffectStatSetsTable         = "GrantedEffectStatSets"
	GrantedEffectStatSetsPerLevelTable = "GrantedEffectStatSetsPerLevel"
	GrantedEffectQualityStatsTable     = "GrantedEffectQualityStats"
	GrantedEffectLabelsTable           = "GrantedEffectLabels"
	GemEffectsTable                    = "GemEffects"
	GemTagsTable                       = "GemTags"
	SkillGemsTable                     = "SkillGems"
	SupportGemsTable                   = "SupportGems"
	SupportGemFamilyTable              = "SupportGemFamily"
	ItemExperiencePerLevelTable        = "ItemExperiencePerLevel"
	CostTypesTable                     = "CostTypes"
	EffectivenessCostConstantsTable    = "EffectivenessCostConstants"
	FlavourTextTable                   = "FlavourText"
	ClientStringsTable                 = "ClientStrings"
)

type ActiveSkillType struct {
	FlagStat *int   `json:"FlagStat"`
	ID       string `json:"Id"`
	Key      int    `json:"_key"`
}

type ActiveSkill struct {
	ID                     string `json:"Id"`
	DisplayedName          string `json:"DisplayedName"` // DisplayName
	Description            string `json:"Description"`
	IconDDSFile            string `json:"Icon_DDSFile"` // Icon
	StatDescription        string `json:"StatDescription"`
	WeaponRequirements     *int   `json:"WeaponRequirements"`     // WeaponRestrictions
	ActiveSkillTypes       []int  `json:"ActiveSkillTypes"`       // SkillTypes
	MinionActiveSkillTypes []int  `json:"MinionActiveSkillTypes"` // MinionSkillTypes
	Key                    int    `json:"_key"`
	SkillTotemID           int    `json:"SkillTotemId"` // SkillTotem
}

type ActiveSkillWeaponRequirement struct {
	ID               string `json:"Id"`
	String           *int   `json:"String"`
	WieldableClasses []int  `json:"WieldableClasses"` // WeaponClass
	Key              int    `json:"_key"`
}

type WieldableClass struct {
	Key       int `json:"_key"`
	ItemClass int `json:"ItemClass"`
}

type ItemClass struct {
	ItemClassCategory *int   `json:"ItemClassCategory"`
	ID                string `json:"Id"`
	Name              string `json:"Name"`
	Key               int    `json:"_key"`
}

type ItemClassCategory struct {
	ID   string `json:"Id"`
	Text string `json:"Text"` // Name
	Key  int    `json:"_key"`
}

type GrantedEffect struct {
	ID                        string `json:"Id"`
	ActiveSkill               *int   `json:"ActiveSkill"`
	AllowedActiveSkillTypes   []int  `json:"AllowedActiveSkillTypes"`  // SupportTypes
	AddedActiveSkillTypes     []int  `json:"AddedActiveSkillTypes"`    // AddTypes
	ExcludedActiveSkillTypes  []int  `json:"ExcludedActiveSkillTypes"` // ExcludeTypes
	AdditionalStatSets        []int  `json:"AdditionalStatSets"`
	CostTypes                 []int  `json:"CostTypes"`                 // CostType
	SupportWeaponRestrictions []int  `json:"SupportWeaponRestrictions"` // WeaponRestrictions
	Key                       int    `json:"_key"`
	CastTime                  int    `json:"CastTime"`
	StatSet                   int    `json:"StatSet"` // GrantedEffectStatSets
	IsSupport                 bool   `json:"IsSupport"`
	SupportsGemsOnly          bool   `json:"SupportsGemsOnly"` // SupportGemsOnly
	CannotBeSupported         bool   `json:"CannotBeSupported"`
	IgnoreMinionTypes         bool   `json:"IgnoreMinionTypes"`
}

type GrantedEffectPerLevel struct {
	CostAmounts                []int   `json:"CostAmounts"`
	ActorLevel                 float64 `json:"ActorLevel"`
	Key                        int     `json:"_key"`
	GrantedEffect              int     `json:"GrantedEffect"`
	Level                      int     `json:"Level"`
	CostMultiplier             int     `json:"CostMultiplier"`
	StoredUses                 int     `json:"StoredUses"`
	Cooldown                   int     `json:"Cooldown"`
	VaalSouls                  int     `json:"VaalSouls"`
	VaalStoredUses             int     `json:"VaalStoredUses"`
	PvPDamageMultiplier        int     `json:"PvPDamageMultiplier"`
	SoulGainPreventionDuration int     `json:"SoulGainPreventionDuration"`
	AttackSpeedMultiplier      int     `json:"AttackSpeedMultiplier"`
	AttackTime                 int     `json:"AttackTime"`
	Reservation                int     `json:"Reservation"`    // SpiritReservation
	EffectOnPlayer             int     `json:"EffectOnPlayer"` // ReservationMultiplier
}

type GrantedEffectStatSet struct {
	ID                             string  `json:"Id"`
	Label                          *int    `json:"Label"` // LabelType
	ImplicitStats                  []int   `json:"ImplicitStats"`
	ConstantStats                  []int   `json:"ConstantStats"`
	ConstantStatsValues            []int   `json:"ConstantStatsValues"`
	IgnoredStats                   []int   `json:"IgnoredStats"` // RemoveStats
	BaseEffectiveness              float64 `json:"BaseEffectiveness"`
	IncrementalEffectiveness       float64 `json:"IncrementalEffectiveness"`
	DamageIncrementalEffectiveness float64 `json:"DamageIncrementalEffectiveness"`
	Key                            int     `json:"_key"`
	UseSetAttackMulti              bool    `json:"Var9"`
}

type GrantedEffectStatSetPerLevel struct {
	BaseResolvedValues    []int     `json:"BaseResolvedValues"`
	AdditionalStatsValues []int     `json:"AdditionalStatsValues"`
	GrantedEffects        []int     `json:"GrantedEffects"`  // GrantedEffect
	AdditionalFlags       []int     `json:"AdditionalFlags"` // AdditionalBooleanStats
	FloatStats            []int     `json:"FloatStats"`
	InterpolationBases    []int     `json:"InterpolationBases"`
	AdditionalStats       []int     `json:"AdditionalStats"`
	StatInterpolations    []int     `json:"StatInterpolations"`
	FloatStatsValues      []float64 `json:"FloatStatsValues"`
	ActorLevel            float64   `json:"ActorLevel"`
	Key                   int       `json:"_key"`
	StatSet               int       `json:"StatSet"` // GrantedEffectStatSets
	GemLevel              int       `json:"GemLevel"`
	SpellCritChance       int       `json:"SpellCritChance"`  // AttackCritChance
	AttackCritChance      int       `json:"AttackCritChance"` // OffhandCritChance
	BaseMultiplier        int       `json:"BaseMultiplier"`
}

type GrantedEffectQualityStat struct {
	Stats                 []int `json:"Stats"`               // GrantedStats
	StatsValuesPermille   []int `json:"StatsValuesPermille"` // StatValues
	ApplyToStatSets       []int `json:"ApplyToStatSets"`     // StatSetIndex
	AltStats              []int `json:"AltStats"`
	AltStatValuesPermille []int `json:"AltStatValuesPermille"` // AltStatValues
	AltApplyToStatSets    []int `json:"AltApplyToStatSets"`    // AltStatSetIndex
	Key                   int   `json:"_key"`
	GrantedEffect         int   `json:"GrantedEffect"`
}

type GrantedEffectLabel struct {
	ID   string `json:"Id"`
	Text string `json:"Text"` // Label
	Key  int    `json:"_key"`
}

type GemEffect struct {
	ID                        string `json:"Id"`
	Name                      string `json:"Name"`
	SupportText               string `json:"SupportText"` // Description
	SupportName               string `json:"SupportName"` // SecondarySupportName
	GemTags                   []int  `json:"GemTags"`     // Tags
	AdditionalGrantedEffects  []int  `json:"AdditionalGrantedEffects"`
	GrantedEffectDisplayOrder []int  `json:"Var12"`
	Key                       int    `json:"_key"`
	GrantedEffect             int    `json:"GrantedEffect"`
}

type GemTag struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Key  int    `json:"_key"`
}

type SkillGem struct {
	GemEffects                     []int `json:"GemEffects"`
	Key                            int   `json:"_key"`
	BaseItemType                   int   `json:"BaseItemType"`
	StrengthRequirementPercent     int   `json:"StrengthRequirementPercent"`     // Str
	DexterityRequirementPercent    int   `json:"DexterityRequirementPercent"`    // Dex
	IntelligenceRequirementPercent int   `json:"IntelligenceRequirementPercent"` // Int
	GemType                        int   `json:"GemType"`                        // IsSupport, see IsSupport
	ItemExperienceType             int   `json:"ItemExperienceType"`             // GemLevelProgression
	CraftingLevel                  int   `json:"CraftingLevel"`                  // Tier
	Tier                           int   `json:"Tier"`                           // SupportTier
	IsVaalVariant                  bool  `json:"IsVaalVariant"`                  // IsVaalGem
}

// IsSupport is PoB's IsSupport: it reads GemType's first byte as a Bool, true only when it is 1.
func (g *SkillGem) IsSupport() bool {
	return g.GemType == 1
}

type SupportGem struct {
	FlavourText *int  `json:"FlavourText"`
	Family      []int `json:"Family"`
	Key         int   `json:"_key"`
	SkillGem    int   `json:"SkillGem"`
	IsLineage   bool  `json:"IsLineage"` // Lineage
}

type SupportGemFamily struct {
	ID   string `json:"Id"`
	Text string `json:"Text"` // Name
	Key  int    `json:"_key"`
}

type ItemExperiencePerLevel struct {
	Key                int `json:"_key"`
	ItemExperienceType int `json:"ItemExperienceType"`
	ItemCurrentLevel   int `json:"ItemCurrentLevel"` // Level
	Experience         int `json:"Experience"`
	Level              int `json:"Level"` // PlayerLevel
}

type CostType struct {
	ID      string `json:"Id"` // Resource
	Key     int    `json:"_key"`
	Stat    int    `json:"Stat"`
	Divisor int    `json:"Divisor"`
}

type EffectivenessCostConstant struct {
	ID         string  `json:"Id"`
	Multiplier float64 `json:"Multiplier"` // Value
	Key        int     `json:"_key"`
}

type FlavourText struct {
	ID   string `json:"Id"`
	Text string `json:"Text"`
	Key  int    `json:"_key"`
}

type ClientString struct {
	ID   string `json:"Id"`
	Text string `json:"Text"`
	Key  int    `json:"_key"`
}
