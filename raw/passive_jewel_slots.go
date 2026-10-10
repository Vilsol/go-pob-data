package raw

type PassiveJewelSlot struct {
	ClusterJewelSize *int  `json:"ClusterJewelSize"`
	ReplacesSlot     *int  `json:"ReplacesSlot"`
	ProxySlot        *int  `json:"ProxySlot"`
	StartIndices     []int `json:"StartIndices"`
	Slot             int   `json:"Slot"`
	ClusterIndex     int   `json:"ClusterIndex"`
	Key              int   `json:"_key"`
}
