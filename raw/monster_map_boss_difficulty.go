package raw

type MonsterMapBossDifficulty struct {
	MapLevel   int `json:"MapLevel"`
	Stat1Value int `json:"Stat1Value"`
	Stat2Value int `json:"Stat2Value"`
	Stat3Value int `json:"Stat3Value"`
	Stat4Value int `json:"Stat4Value"`
	Stat5Value int `json:"Stat5Value"`
	StatsKey1  int `json:"StatsKey1"`
	StatsKey2  int `json:"StatsKey2"`
	StatsKey3  int `json:"StatsKey3"`
	StatsKey4  int `json:"StatsKey4"`
	StatsKey5  int `json:"StatsKey5"`
	Key        int `json:"_key"`
}
