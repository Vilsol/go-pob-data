package raw

type GrantedEffect struct {
	ActiveSkill           *int   `json:"ActiveSkill"`
	PlusVersionOf         *int   `json:"RegularVariant"`
	Animation             *int   `json:"Animation"`
	SupportGemLetter      string `json:"SupportGemLetter"`
	ID                    string `json:"Id"`
	SupportTypes          []int  `json:"AllowedActiveSkillTypes"`
	AddTypes              []int  `json:"AddedActiveSkillTypes"`
	ExcludeTypes          []int  `json:"ExcludedActiveSkillTypes"`
	WeaponRestrictions    []int  `json:"SupportWeaponRestrictions"`
	AddMinionTypes        []int  `json:"AddedMinionActiveSkillTypes"`
	Attribute             int    `json:"Attribute"`
	CastTime              int    `json:"CastTime"`
	GrantedEffectStatSets int    `json:"StatSet"`
	Key                   int    `json:"_key"`
	IgnoreMinionTypes     bool   `json:"IgnoreMinionTypes"`
	CannotBeSupported     bool   `json:"CannotBeSupported"`
	SupportsGemsOnly      bool   `json:"SupportsGemsOnly"`
	IsSupport             bool   `json:"IsSupport"`
}
