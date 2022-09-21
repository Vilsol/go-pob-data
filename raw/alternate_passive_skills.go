package raw

type AlternatePassiveSkill struct {
	DDSIcon                  string        `json:"DDSIcon"`
	FlavourText              string        `json:"FlavourText"`
	ID                       string        `json:"Id"`
	Name                     string        `json:"Name"`
	AchievementItemsKeys     []interface{} `json:"AchievementItemsKeys"`
	StatsKeys                []int         `json:"StatsKeys"`
	PassiveType              []int         `json:"PassiveType"`
	RandomMin                int           `json:"RandomMin"`
	RandomMax                int           `json:"RandomMax"`
	SpawnWeight              int           `json:"SpawnWeight"`
	Stat1Max                 int           `json:"Stat1Max"`
	Stat1Min                 int           `json:"Stat1Min"`
	Stat2Max                 int           `json:"Stat2Max"`
	Stat2Min                 int           `json:"Stat2Min"`
	AlternateTreeVersionsKey int           `json:"AlternateTreeVersionsKey"`
	Key                      int           `json:"_key"`
}
