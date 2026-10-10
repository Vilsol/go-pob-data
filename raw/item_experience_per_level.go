package raw

type ItemExperiencePerLevel struct {
	// ItemExperienceType groups a progression's levels (3.21+), as SkillGem.GemLevelProgression refers to it.
	ItemExperienceType *int `json:"ItemExperienceType"`
	BaseItemTypesKey   int  `json:"BaseItemTypesKey"`
	Experience         int  `json:"Experience"`
	ItemCurrentLevel   int  `json:"ItemCurrentLevel"`
	Key                int  `json:"_key"`
}
