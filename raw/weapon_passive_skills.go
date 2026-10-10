package raw

type WeaponPassiveSkill struct {
	ID                string `json:"Id"`
	Icon              string `json:"Icon"`
	NodeSpawnLocation []int  `json:"NodeSpawnLocation"`
	Tags              []int  `json:"Tags"`
	Achievements      []int  `json:"Achievements"`
	Tier              int    `json:"Tier"`
	Mod               int    `json:"Mod"`
	Type              int    `json:"Type"`
	Key               int    `json:"_key"`
}
