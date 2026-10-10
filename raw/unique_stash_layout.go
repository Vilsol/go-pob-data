package raw

type UniqueStashLayout struct {
	RenamedVersion             *int `json:"RenamedVersion"`
	BaseVersion                *int `json:"BaseVersion"`
	WordsKey                   int  `json:"WordsKey"`
	ItemVisualIdentityKey      int  `json:"ItemVisualIdentityKey"`
	UniqueStashTypesKey        int  `json:"UniqueStashTypesKey"`
	OverrideWidth              int  `json:"OverrideWidth"`
	OverrideHeight             int  `json:"OverrideHeight"`
	Key                        int  `json:"_key"`
	ShowIfEmptyChallengeLeague bool `json:"ShowIfEmptyChallengeLeague"`
	ShowIfEmptyStandard        bool `json:"ShowIfEmptyStandard"`
	IsAlternateArt             bool `json:"IsAlternateArt"`
}
