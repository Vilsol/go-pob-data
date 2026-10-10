package raw

type GemEffect struct {
	GrantedEffectHardmode  *int   `json:"GrantedEffectHardmode"`
	GrantedEffect2         *int   `json:"GrantedEffect2"`
	GrantedEffect2Hardmode *int   `json:"GrantedEffect2Hardmode"`
	ConsumedMod            *int   `json:"Consumed_ModsKey"`
	ID                     string `json:"Id"`
	Name                   string `json:"Name"`
	SupportText            string `json:"SupportText"`
	SupportName            string `json:"SupportName"`
	GemTags                []int  `json:"GemTags"`
	GrantedEffect          int    `json:"GrantedEffect"`
	ItemColor              int    `json:"ItemColor"`
	Key                    int    `json:"_key"`
}
