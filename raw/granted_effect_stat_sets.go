package raw

type GrantedEffectStatSet struct {
	ID                       string  `json:"Id"`
	ImplicitStats            []int   `json:"ImplicitStats"`
	ConstantStats            []int   `json:"ConstantStats"`
	ConstantStatsValues      []int   `json:"ConstantStatsValues"`
	Key                      int     `json:"_key"`
	BaseEffectiveness        float64 `json:"BaseEffectiveness"`
	IncrementalEffectiveness float64 `json:"IncrementalEffectiveness"`
}
