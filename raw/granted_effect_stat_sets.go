package raw

type GrantedEffectStatSet struct {
	Key                      int     `json:"_key"`
	ID                       string  `json:"Id"`
	ImplicitStats            []int   `json:"ImplicitStats"`
	ConstantStats            []int   `json:"ConstantStats"`
	ConstantStatsValues      []int   `json:"ConstantStatsValues"`
	BaseEffectiveness        float64 `json:"BaseEffectiveness"`
	IncrementalEffectiveness float64 `json:"IncrementalEffectiveness"`
}
