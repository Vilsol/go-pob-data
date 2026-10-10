package raw

type BuffTemplate struct {
	BuffVisualsKey     *int   `json:"BuffVisualsKey"`
	ID                 string `json:"Id"`
	BuffStatValues     []int  `json:"Buff_StatValues"`
	StatsKey           []int  `json:"StatsKey"`
	BuffDefinitionsKey int    `json:"BuffDefinitionsKey"`
	AuraRadius         int    `json:"AuraRadius"`
	Key                int    `json:"_key"`
}
