package raw

type PassiveSkillTattooTargetSet struct {
	ID        string `json:"Id"`
	Name      string `json:"Name"`
	Qualifier string `json:"Qualifier"`
	Set       []int  `json:"Set"`
	Key       int    `json:"_key"`
}
