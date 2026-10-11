package raw

type MonsterResistance struct {
	ID                 string `json:"Id"`
	FireNormal         int    `json:"FireNormal"`
	ColdNormal         int    `json:"ColdNormal"`
	LightningNormal    int    `json:"LightningNormal"`
	ChaosNormal        int    `json:"ChaosNormal"`
	FireCruel          int    `json:"FireCruel"`
	ColdCruel          int    `json:"ColdCruel"`
	LightningCruel     int    `json:"LightningCruel"`
	ChaosCruel         int    `json:"ChaosCruel"`
	FireMerciless      int    `json:"FireMerciless"`
	ColdMerciless      int    `json:"ColdMerciless"`
	LightningMerciless int    `json:"LightningMerciless"`
	ChaosMerciless     int    `json:"ChaosMerciless"`
	Key                int    `json:"_key"`
}
