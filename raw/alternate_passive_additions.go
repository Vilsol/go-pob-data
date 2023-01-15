package raw

type AlternatePassiveAddition struct {
	ID                       string `json:"Id"`
	PassiveType              []int  `json:"PassiveType"`
	StatsKeys                []int  `json:"StatsKeys"`
	AlternateTreeVersionsKey int    `json:"AlternateTreeVersionsKey"`
	SpawnWeight              int    `json:"SpawnWeight"`
	Stat1Max                 int    `json:"Stat1Max"`
	Stat1Min                 int    `json:"Stat1Min"`
	Stat2Min                 int    `json:"Var6"`
	Stat2Max                 int    `json:"Var7"`
	Key                      int    `json:"_key"`
}
