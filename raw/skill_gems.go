package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type SkillGem struct {
	VaalGem                *int `json:"VaalVariant_BaseItemTypesKey"`
	RegularVariant         *int `json:"RegularVariant"`
	AwakenedVariant        *int `json:"AwakenedVariant"`
	GlobalGemLevelStat     *int `json:"MinionGlobalSkillLevelStat"`
	SecondaryGrantedEffect *int `json:"GrantedEffectsKey2"`
	HungryLoopMod          *int `json:"Consumed_ModsKey"`
	// GemLevelProgression is the ItemExperienceType of the gem's ItemExperiencePerLevel rows (3.21+).
	GemLevelProgression  *int   `json:"ItemExperienceType"`
	SecondarySupportName string `json:"SupportSkillName"`
	Description          string `json:"Description"`
	Tags                 []int  `json:"GemTagsKeys"`
	// GemVariants are the gem's GemEffects (3.28+).
	GemVariants   []int `json:"GemVariants"`
	Int           int   `json:"Int"`
	Dex           int   `json:"Dex"`
	BaseItemType  int   `json:"BaseItemTypesKey"`
	Str           int   `json:"Str"`
	GrantedEffect int   `json:"GrantedEffectsKey"`
	GemColour     int   `json:"GemColour"`
	Key           int   `json:"_key"`
	IsVaalGem     bool  `json:"IsVaalVariant"`
	IsSupport     bool  `json:"IsSupport"`
}

// UnmarshalJSON reads Str, Dex and Int under their 3.24+ schema names too.
func (s *SkillGem) UnmarshalJSON(b []byte) error {
	type plain SkillGem
	var named struct {
		Str *int `json:"StrengthRequirementPercent"`
		Dex *int `json:"DexterityRequirementPercent"`
		Int *int `json:"IntelligenceRequirementPercent"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode skill gem")
	}
	*s = SkillGem(named.plain)
	for _, f := range []struct {
		from *int
		to   *int
	}{{named.Str, &s.Str}, {named.Dex, &s.Dex}, {named.Int, &s.Int}} {
		if f.from != nil {
			*f.to = *f.from
		}
	}
	return nil
}
