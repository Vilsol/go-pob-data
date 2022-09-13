package raw

type AlternatePassiveAddition struct {
	AlternateTreeVersionsKey int    `json:"AlternateTreeVersionsKey"`
	ID                       string `json:"Id"`
	PassiveType              []int  `json:"PassiveType"`
	SpawnWeight              int    `json:"SpawnWeight"`
	Stat1Max                 int    `json:"Stat1Max"`
	Stat1Min                 int    `json:"Stat1Min"`
	StatsKeys                []int  `json:"StatsKeys"`
	Key                      int    `json:"_key"`
}
