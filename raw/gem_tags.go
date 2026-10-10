package raw

type GemTag struct {
	Stat1 *int   `json:"Stat1"`
	Stat2 *int   `json:"Stat2"`
	Stat3 *int   `json:"Stat3"`
	Stat4 *int   `json:"Stat4"`
	ID    string `json:"Id"`
	Tag   string `json:"Tag"`
	Key   int    `json:"_key"`
}
