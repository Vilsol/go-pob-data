package raw

type Tincture struct {
	BaseItem       int `json:"BaseItem"`
	DebuffInterval int `json:"DebuffInterval"`
	Cooldown       int `json:"Cooldown"`
	Key            int `json:"_key"`
}
