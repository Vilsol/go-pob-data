package treegen

// PoE2 rows, named as go-pob-data's schema names them; the comment beside a column is the name
// PathOfBuilding-PoE2's src/Export/spec.lua gives it, which passivetree.lua reads.

//nolint:govet // fields follow the export key order
type passiveSkill2 struct {
	ID                     string `json:"Id"`
	Icon                   string `json:"Icon_DDSFile"`
	Stats                  []int  `json:"Stats"`
	Stat1Value             int    `json:"Stat1Value"`
	Stat2Value             int    `json:"Stat2Value"`
	Stat3Value             int    `json:"Stat3Value"`
	Stat4Value             int    `json:"Stat4Value"`
	PassiveSkillGraphID    uint32 `json:"PassiveSkillGraphId"`
	Name                   string `json:"Name"`
	Characters             []int  `json:"Characters"` // ClassStart
	IsKeystone             bool   `json:"IsKeystone"`
	IsNotable              bool   `json:"IsNotable"`
	FlavourText            string `json:"FlavourText"`
	IsJustIcon             bool   `json:"IsJustIcon"` // IsOnlyImage
	IsJewelSocket          bool   `json:"IsJewelSocket"`
	Ascendancy             *int   `json:"Ascendancy"`
	IsAscendancyStart      bool   `json:"IsAscendancyStartingNode"`
	SkillPointsGranted     int    `json:"SkillPointsGranted"` // PassivePointsGranted
	IsMultipleChoice       bool   `json:"IsMultipleChoice"`
	IsMultipleChoiceOption bool   `json:"IsMultipleChoiceOption"`
	Stat5Value             int    `json:"Stat5Value"`
	IsAnointmentOnly       bool   `json:"IsAnointmentOnly"` // AnointOnly
	MasteryGroup           *int   `json:"MasteryGroup"`
	IsAttribute            bool   `json:"IsAttribute"`
	GrantedSkill           *int   `json:"GrantedSkill"`
	WeaponPointsGranted    int    `json:"WeaponPointsGranted"`
	IsFree                 bool   `json:"IsFree"`               // FreeAllocate
	ApplyToArmour          bool   `json:"Var51"`                // ApplyToArmour?
	UnlockedBy             []int  `json:"UnlockedBy"`           // ConstraintNode
	VisibleForAscendancy   *int   `json:"VisibleForAscendancy"` // AscendancyUnlock
	NodeFrameArt           *int   `json:"NodeFrameArt"`         // FrameArt
	Stat6Value             int    `json:"Stat6Value"`
	Stat7Value             int    `json:"Stat7Value"`
}

func (p *passiveSkill2) statValues() []int {
	return []int{p.Stat1Value, p.Stat2Value, p.Stat3Value, p.Stat4Value, p.Stat5Value, p.Stat6Value, p.Stat7Value}
}

//nolint:govet // fields follow the export key order
type character2 struct {
	Name             string `json:"Name"`
	IntegerID        int    `json:"IntegerId"`
	BaseStrength     int    `json:"BaseStrength"`
	BaseDexterity    int    `json:"BaseDexterity"`
	BaseIntelligence int    `json:"BaseIntelligence"`
}

//nolint:govet // fields follow the export key order
type ascendancy2 struct {
	ID               string `json:"Id"`
	Character        *int   `json:"Character"` // Class
	Name             string `json:"Name"`
	TreeRegionVector int    `json:"TreeRegionVector"` // distanceTree
	Disabled         bool   `json:"Disabled"`         // isDisabled
	BaseAscendancy   *int   `json:"BaseAscendancy"`   // Replace
	UIArt            *int   `json:"UIArt"`
}

type treeUIArt2 struct {
	JewelFrame *int `json:"JewelFrame"`
}

type nodeFrameArt2 struct {
	Normal      string `json:"Normal"`
	CanAllocate string `json:"CanAllocate"`
	Active      string `json:"Active"`
}

type masteryGroup2 struct {
	Art *int `json:"Art"` // MasteryArt
}

type masteryArt2 struct {
	ActiveEffectImage string `json:"ActiveEffectImage"` // Effect
}

type classOverride2 struct {
	Character *int `json:"CharacterToOverrideFor"`
	Original  *int `json:"SkillToOverride"`
	Switched  *int `json:"Override"`
}

type ascendancyOverride2 struct {
	Original *int `json:"SkillToOverride"`
	Switched *int `json:"Override"`
}

type blightResult2 struct {
	PassiveSkill *int `json:"PassiveSkill"`
}

type blightRecipe2 struct {
	Result *int  `json:"BlightCraftingResult"`
	Items  []int `json:"BlightCraftingItems"`
}

type blightItem2 struct {
	NameShort string `json:"NameShort"`
}

type skillGem2 struct {
	BaseItemType *int  `json:"BaseItemType"`
	GemEffects   []int `json:"GemEffects"`
}
