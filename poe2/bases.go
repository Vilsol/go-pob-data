package poe2

// The tables PoB-PoE2's base exporter (src/Export/Scripts/bases.lua) reads besides the skill tables, and
// SkillGemSupports (skillGemList.lua). Mods is not typed: PoB's spec and the json's columns diverge in
// order past Stat6. Columns keep the names the json carries; where PoB's spec names a column differently,
// the PoB name follows in a comment.
const (
	ArmourTypesTable           = "ArmourTypes"
	ShieldTypesTable           = "ShieldTypes"
	WeaponTypesTable           = "WeaponTypes"
	FlasksTable                = "Flasks"
	ComponentChargesTable      = "ComponentCharges"
	UtilityFlaskBuffsTable     = "UtilityFlaskBuffs"
	BuffDefinitionsTable       = "BuffDefinitions"
	AttributeRequirementsTable = "AttributeRequirements"
	ItemSpiritTable            = "ItemSpirit"
	ItemInherentSkillsTable    = "ItemInherentSkills"
	ItemExperienceTypesTable   = "ItemExperienceTypes"
	TagsTable                  = "Tags"
	SkillGemSupportsTable      = "SkillGemSupports"
)

type ArmourType struct {
	Key                    int `json:"_key"`
	BaseItemType           int `json:"BaseItemType"`
	Armour                 int `json:"Armour"`
	Evasion                int `json:"Evasion"`
	EnergyShield           int `json:"EnergyShield"`
	IncreasedMovementSpeed int `json:"IncreasedMovementSpeed"` // MovementPenalty
	Ward                   int `json:"Ward"`
}

type ShieldType struct {
	Key          int `json:"_key"`
	BaseItemType int `json:"BaseItemType"`
	Block        int `json:"Block"`
}

type WeaponType struct {
	Key          int `json:"_key"`
	BaseItemType int `json:"BaseItemType"`
	CritChance   int `json:"CritChance"`
	WeaponClass  int `json:"WeaponClass"`
	Speed        int `json:"Speed"`
	DamageMin    int `json:"DamageMin"`
	DamageMax    int `json:"DamageMax"`
	RangeMax     int `json:"RangeMax"` // Range
	ReloadTime   int `json:"ReloadTime"`
}

type Flask struct {
	BuffDefinition *int   `json:"BuffDefinition"` // Buff
	Name           string `json:"Name"`
	UtilityBuff    []int  `json:"UtilityBuff"` // UtilityBuffs
	Key            int    `json:"_key"`
	BaseItemType   int    `json:"BaseItemType"`
	Type           int    `json:"Type"`
	LifePerUse     int    `json:"LifePerUse"`
	ManaPerUse     int    `json:"ManaPerUse"`
	RecoveryTime   int    `json:"RecoveryTime"`
}

// ComponentCharge is keyed by the base's Id, not its row.
type ComponentCharge struct {
	BaseItemTypesKey string `json:"BaseItemTypesKey"` // BaseItemType
	Key              int    `json:"_key"`
	MaxCharges       int    `json:"MaxCharges"` // Max
	PerCharge        int    `json:"PerCharge"`  // PerUse
}

type UtilityFlaskBuff struct {
	StatValues     []int `json:"StatValues"`
	Key            int   `json:"_key"`
	BuffDefinition int   `json:"BuffDefinition"` // BuffDefinitionsKey
}

type BuffDefinition struct {
	ID           string `json:"Id"`
	Name         string `json:"Name"`
	Stats        []int  `json:"Stats"`
	GrantedFlags []int  `json:"GrantedFlags"`
	GrantedStats []int  `json:"GrantedStats"`
	Key          int    `json:"_key"`
}

type AttributeRequirement struct {
	Key          int `json:"_key"`
	BaseItemType int `json:"BaseItemType"` // BaseType
	ReqStr       int `json:"ReqStr"`
	ReqDex       int `json:"ReqDex"`
	ReqInt       int `json:"ReqInt"`
}

type ItemSpirit struct {
	Key           int `json:"_key"`
	BaseItemType  int `json:"BaseItemType"`
	SpiritGranted int `json:"SpiritGranted"` // Value
}

type ItemInherentSkill struct {
	SkillsGranted []int `json:"SkillsGranted"` // Skill
	Key           int   `json:"_key"`
	BaseItemType  int   `json:"BaseItemType"`
	IsWeapon      bool  `json:"IsWeapon"` // NoReservation
}

// ItemExperienceType resolves SkillGem.ItemExperienceType.
type ItemExperienceType struct {
	ID  string `json:"Id"`
	Key int    `json:"_key"`
}

// Tag resolves BaseItemType.Tags.
type Tag struct {
	ID  string `json:"Id"`
	Key int    `json:"_key"`
}

type SkillGemSupport struct {
	Supports []int `json:"Supports"` // SuggestedSupport
	Key      int   `json:"_key"`
	SkillGem int   `json:"SkillGem"` // ActiveGem
}
