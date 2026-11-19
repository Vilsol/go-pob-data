package raw

type MonsterPackEntry struct {
	MonsterPacksKey     *int   `json:"MonsterPacksKey"`
	MonsterVarietiesKey *int   `json:"MonsterVarietiesKey"`
	ID                  string `json:"Id"`
	Weight              int    `json:"Weight"`
	Key                 int    `json:"_key"`
	Flag                bool   `json:"Flag"`
}
