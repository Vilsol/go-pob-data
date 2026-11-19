package raw

type Map struct {
	BaseItemTypesKey     *int  `json:"BaseItemTypesKey"`
	RegularWorldAreasKey *int  `json:"Regular_WorldAreasKey"`
	UniqueWorldAreasKey  *int  `json:"Unique_WorldAreasKey"`
	MonsterPacksKeys     []int `json:"MonsterPacksKeys"`
	Tier                 int   `json:"Tier"`
	Key                  int   `json:"_key"`
}
