package raw

type MonsterType struct {
	Resistances                  *int   `json:"Resistances"`
	ID                           string `json:"Id"`
	Accuracy                     int    `json:"Accuracy"`
	Armour                       int    `json:"Armour"`
	Evasion                      int    `json:"Evasion"`
	EnergyShield                 int    `json:"EnergyShield"`
	DamageSpread                 int    `json:"DamageSpread"`
	Key                          int    `json:"_key"`
	IsPlayerMinion               bool   `json:"IsPlayerMinion"`
	AltLife1                     bool   `json:"AltLife1"`
	AltLife2                     bool   `json:"AltLife2"`
	BaseDamageIgnoresAttackSpeed bool   `json:"BaseDamageIgnoresAttackSpeed"`
}
