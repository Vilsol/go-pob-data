package raw

type ComponentCharge struct {
	BaseItemTypesKey string `json:"BaseItemTypesKey"`
	MaxCharges       int    `json:"MaxCharges"`
	MaxCharges2      int    `json:"MaxCharges2"`
	PerCharge        int    `json:"PerCharge"`
	PerCharge2       int    `json:"PerCharge2"`
	Key              int    `json:"_key"`
}
