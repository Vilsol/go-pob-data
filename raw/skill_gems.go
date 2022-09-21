package raw

type SkillGem struct {
	VaalGem                *int   `json:"VaalVariant_BaseItemTypesKey"`
	RegularVariant         *int   `json:"RegularVariant"`
	AwakenedVariant        *int   `json:"AwakenedVariant"`
	GlobalGemLevelStat     *int   `json:"MinionGlobalSkillLevelStat"`
	SecondaryGrantedEffect *int   `json:"GrantedEffectsKey2"`
	HungryLoopMod          *int   `json:"Consumed_ModsKey"`
	SecondarySupportName   string `json:"SupportSkillName"`
	Description            string `json:"Description"`
	Tags                   []int  `json:"GemTagsKeys"`
	Int                    int    `json:"Int"`
	Dex                    int    `json:"Dex"`
	BaseItemType           int    `json:"BaseItemTypesKey"`
	Str                    int    `json:"Str"`
	GrantedEffect          int    `json:"GrantedEffectsKey"`
	Key                    int    `json:"_key"`
	IsVaalGem              bool   `json:"IsVaalVariant"`
}
