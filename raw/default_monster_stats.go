package raw

type DefaultMonsterStat struct {
	Accuracy     int     `json:"Accuracy"`
	AllyLife     int     `json:"AllyLife"`
	Armour       int     `json:"Armour"`
	Damage       float64 `json:"Damage"`
	Damage2      float64 `json:"Damage2"`
	Difficulty   int     `json:"Difficulty"`
	DisplayLevel string  `json:"DisplayLevel"`
	Evasion      int     `json:"Evasion"`
	Experience   int     `json:"Experience"`
	Life         int     `json:"Life"`
	Key          int     `json:"_key"`
}
