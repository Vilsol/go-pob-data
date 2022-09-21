package raw

type GrantedEffectQualityStat struct {
	StatsKeys           []int `json:"StatsKeys"`
	StatsValuesPermille []int `json:"StatsValuesPermille"`
	GrantedEffectsKey   int   `json:"GrantedEffectsKey"`
	SetID               int   `json:"SetId"`
	Weight              int   `json:"Weight"`
	Key                 int   `json:"_key"`
}
