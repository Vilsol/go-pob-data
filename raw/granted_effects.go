package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

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
	StatSet2              []int  `json:"StatSet2"` // further GrantedEffectStatSets
	Attribute             int    `json:"Attribute"`
	CastTime              int    `json:"CastTime"`
	GrantedEffectStatSets int    `json:"StatSet1"`
	Key                   int    `json:"_key"`
	IgnoreMinionTypes     bool   `json:"IgnoreMinionTypes"`
	CannotBeSupported     bool   `json:"CannotBeSupported"`
	SupportsGemsOnly      bool   `json:"SupportsGemsOnly"`
	IsSupport             bool   `json:"IsSupport"`
}

// UnmarshalJSON reads the stat set columns under the names tables extracted before 3.28 give them:
// StatSet1 as StatSet, StatSet2 as Var27.
func (g *GrantedEffect) UnmarshalJSON(b []byte) error {
	type plain GrantedEffect
	var named struct {
		StatSet *int  `json:"StatSet"`
		Var27   []int `json:"Var27"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode granted effect")
	}
	*g = GrantedEffect(named.plain)
	setFrom(&g.GrantedEffectStatSets, named.StatSet)
	if named.Var27 != nil {
		g.StatSet2 = named.Var27
	}
	return nil
}
