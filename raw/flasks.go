package raw

type Flask struct {
	BuffDefinitionsKey *int   `json:"BuffDefinitionsKey"`
	Name               string `json:"Name"`
	BuffStatValues     []int  `json:"BuffStatValues"`
	BuffStatValues2    []int  `json:"BuffStatValues2"`
	BaseItemTypesKey   int    `json:"BaseItemTypesKey"`
	Group              int    `json:"Group"`
	LifePerUse         int    `json:"LifePerUse"`
	ManaPerUse         int    `json:"ManaPerUse"`
	RecoveryTime       int    `json:"RecoveryTime"`
	RecoveryTime2      int    `json:"RecoveryTime2"`
	Key                int    `json:"_key"`
}
