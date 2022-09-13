package raw

type GrantedEffect struct {
	ID                    string `json:"Id"`
	IsSupport             bool   `json:"IsSupport"`
	SupportTypes          []int  `json:"AllowedActiveSkillTypes"`
	SupportGemLetter      string `json:"SupportGemLetter"`
	Attribute             int    `json:"Attribute"`
	AddTypes              []int  `json:"AddedActiveSkillTypes"`
	ExcludeTypes          []int  `json:"ExcludedActiveSkillTypes"`
	SupportsGemsOnly      bool   `json:"SupportsGemsOnly"`
	CannotBeSupported     bool   `json:"CannotBeSupported"`
	CastTime              int    `json:"CastTime"`
	ActiveSkill           *int   `json:"ActiveSkill"`
	IgnoreMinionTypes     bool   `json:"IgnoreMinionTypes"`
	AddMinionTypes        []int  `json:"AddedMinionActiveSkillTypes"`
	Animation             *int   `json:"Animation"`
	WeaponRestrictions    []int  `json:"SupportWeaponRestrictions"`
	PlusVersionOf         *int   `json:"RegularVariant"`
	GrantedEffectStatSets int    `json:"StatSet"`
	Key                   int    `json:"_key"`
}
