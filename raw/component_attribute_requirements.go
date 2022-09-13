package raw

type ComponentAttributeRequirement struct {
	BaseItemTypesKey string `json:"BaseItemTypesKey"`
	ReqDex           int    `json:"ReqDex"`
	ReqInt           int    `json:"ReqInt"`
	ReqStr           int    `json:"ReqStr"`
	Key              int    `json:"_key"`
}
