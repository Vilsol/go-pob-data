package raw

type BuffDefinition struct {
	MaximumStatsKey *int   `json:"Maximum_StatsKey"`
	CurrentStatsKey *int   `json:"Current_StatsKey"`
	MinStat         *int   `json:"MinStat"`
	ID              string `json:"Id"`
	Description     string `json:"Description"`
	Name            string `json:"Name"`
	StatsKeys       []int  `json:"StatsKeys"`
	Stats2          []int  `json:"Stats2"`
	GrantedFlags    []int  `json:"GrantedFlags"`
	GrantedStats    []int  `json:"GrantedStats"`
	ConditionStats  []int  `json:"ConditionStats"`
	Key             int    `json:"_key"`
}
