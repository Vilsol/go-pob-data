package raw

type GrantedEffectStatSetsPerLevel struct {
	GrantedEffects         []int     `json:"GrantedEffects"`
	AdditionalStats        []int     `json:"AdditionalStats"`
	AdditionalStatsValues  []int     `json:"AdditionalStatsValues"`
	StatInterpolations     []int     `json:"StatInterpolations"`
	AdditionalBooleanStats []int     `json:"AdditionalFlags"`
	BaseResolvedValues     []int     `json:"BaseResolvedValues"`
	InterpolationBases     []int     `json:"InterpolationBases"`
	FloatStats             []int     `json:"FloatStats"`
	FloatStatsValues       []float64 `json:"FloatStatsValues"`
	BaseMultiplier         int       `json:"BaseMultiplier"`
	GemLevel               int       `json:"GemLevel"`
	DamageEffectiveness    int       `json:"DamageEffectiveness"`
	PlayerLevelReq         int       `json:"PlayerLevelReq"`
	OffhandCritChance      int       `json:"AttackCritChance"`
	AttackCritChance       int       `json:"SpellCritChance"`
	StatSet                int       `json:"StatSet"`
	Key                    int       `json:"_key"`
}
