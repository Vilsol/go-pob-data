package raw

type Flask struct {
	BaseItemTypesKey   int    `json:"BaseItemTypesKey"`
	BuffDefinitionsKey *int   `json:"BuffDefinitionsKey"`
	BuffStatValues     []int  `json:"BuffStatValues"`
	BuffStatValues2    []int  `json:"BuffStatValues2"`
	Group              int    `json:"Group"`
	LifePerUse         int    `json:"LifePerUse"`
	ManaPerUse         int    `json:"ManaPerUse"`
	Name               string `json:"Name"`
	RecoveryTime       int    `json:"RecoveryTime"`
	RecoveryTime2      int    `json:"RecoveryTime2"`
	Key                int    `json:"_key"`
}
