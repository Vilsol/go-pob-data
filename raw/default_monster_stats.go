package raw

type DefaultMonsterStat struct {
	DisplayLevel string  `json:"DisplayLevel"`
	Accuracy     int     `json:"Accuracy"`
	AllyLife     int     `json:"AllyLife"`
	Armour       int     `json:"Armour"`
	Damage       float64 `json:"Damage"`
	Damage2      float64 `json:"Damage2"`
	Difficulty   int     `json:"Difficulty"`
	Evasion      int     `json:"Evasion"`
	Experience   int     `json:"Experience"`
	Life         int     `json:"Life"`
	Key          int     `json:"_key"`
}
