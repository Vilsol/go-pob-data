package raw

type EffectivenessCostConstant struct {
	ID         string  `json:"Id"`
	Multiplier float64 `json:"Multiplier"` // PoB's Value
	Key        int     `json:"_key"`
}
