package raw

type GrantedEffectQualityStat struct {
	GrantedEffectsKey   int   `json:"GrantedEffectsKey"`
	SetID               int   `json:"SetId"`
	StatsKeys           []int `json:"StatsKeys"`
	StatsValuesPermille []int `json:"StatsValuesPermille"`
	Weight              int   `json:"Weight"`
	Key                 int   `json:"_key"`
}
