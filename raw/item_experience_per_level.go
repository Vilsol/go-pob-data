package raw

type ItemExperiencePerLevel struct {
	BaseItemTypesKey int `json:"BaseItemTypesKey"`
	Experience       int `json:"Experience"`
	ItemCurrentLevel int `json:"ItemCurrentLevel"`
	Key              int `json:"_key"`
}
