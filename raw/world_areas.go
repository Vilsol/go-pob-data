package raw

type WorldArea struct {
	ID                           string `json:"Id"`
	Name                         string `json:"Name"`
	BossesMonsterVarietiesKeys   []int  `json:"Bosses_MonsterVarietiesKeys"`
	MonstersMonsterVarietiesKeys []int  `json:"Monsters_MonsterVarietiesKeys"`
	ModsKeys                     []int  `json:"ModsKeys"`
	AreaTypeTags                 []int  `json:"AreaTypeTags"`
	Tags                         []int  `json:"Tags"`
	Act                          int    `json:"Act"`
	AreaLevel                    int    `json:"AreaLevel"`
	MaxLevel                     int    `json:"MaxLevel"`
	Key                          int    `json:"_key"`
	IsTown                       bool   `json:"IsTown"`
	IsHideout                    bool   `json:"IsHideout"`
	IsMapArea                    bool   `json:"IsMapArea"`
	IsUniqueMapArea              bool   `json:"IsUniqueMapArea"`
	IsVaalArea                   bool   `json:"IsVaalArea"`
	IsLabyrinthArea              bool   `json:"IsLabyrinthArea"`
}
