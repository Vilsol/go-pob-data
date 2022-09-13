package raw

type GrantedEffectStatSetsPerLevel struct {
	AdditionalBooleanStats []int     `json:"AdditionalFlags"`
	AdditionalStats        []int     `json:"AdditionalStats"`
	AdditionalStatsValues  []int     `json:"AdditionalStatsValues"`
	AttackCritChance       int       `json:"SpellCritChance"`
	BaseMultiplier         int       `json:"BaseMultiplier"`
	BaseResolvedValues     []int     `json:"BaseResolvedValues"`
	DamageEffectiveness    int       `json:"DamageEffectiveness"`
	FloatStats             []int     `json:"FloatStats"`
	FloatStatsValues       []float64 `json:"FloatStatsValues"`
	GemLevel               int       `json:"GemLevel"`
	GrantedEffects         []int     `json:"GrantedEffects"`
	InterpolationBases     []int     `json:"InterpolationBases"`
	PlayerLevelReq         int       `json:"PlayerLevelReq"`
	OffhandCritChance      int       `json:"AttackCritChance"`
	StatInterpolations     []int     `json:"StatInterpolations"`
	StatSet                int       `json:"StatSet"`
	Key                    int       `json:"_key"`
}
