package raw

type MonsterPack struct {
	ID                              string `json:"Id"`
	WorldAreasKeys                  []int  `json:"WorldAreasKeys"`
	BossMonsterMonsterVarietiesKeys []int  `json:"BossMonster_MonsterVarietiesKeys"`
	Tags                            []int  `json:"Tags"`
	MinCount                        int    `json:"MinCount"`
	MaxCount                        int    `json:"MaxCount"`
	BossMonsterSpawnChance          int    `json:"BossMonsterSpawnChance"`
	BossMonsterCount                int    `json:"BossMonsterCount"`
	MinLevel                        int    `json:"MinLevel"`
	MaxLevel                        int    `json:"MaxLevel"`
	Key                             int    `json:"_key"`
}
