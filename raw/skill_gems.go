package raw

type SkillGem struct {
	BaseItemType           int    `json:"BaseItemTypesKey"`
	GrantedEffect          int    `json:"GrantedEffectsKey"`
	Str                    int    `json:"Str"`
	Dex                    int    `json:"Dex"`
	Int                    int    `json:"Int"`
	Tags                   []int  `json:"GemTagsKeys"`
	VaalGem                *int   `json:"VaalVariant_BaseItemTypesKey"`
	IsVaalGem              bool   `json:"IsVaalVariant"`
	Description            string `json:"Description"`
	HungryLoopMod          *int   `json:"Consumed_ModsKey"`
	SecondaryGrantedEffect *int   `json:"GrantedEffectsKey2"`
	GlobalGemLevelStat     *int   `json:"MinionGlobalSkillLevelStat"`
	SecondarySupportName   string `json:"SupportSkillName"`
	AwakenedVariant        *int   `json:"AwakenedVariant"`
	RegularVariant         *int   `json:"RegularVariant"`
	Key                    int    `json:"_key"`
}
