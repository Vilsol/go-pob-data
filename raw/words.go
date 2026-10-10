package raw

type Word struct {
	Text              string `json:"Text"`
	Text2             string `json:"Text2"`
	Inflection        string `json:"Inflection"`
	SpawnWeightTags   []int  `json:"SpawnWeight_Tags"`
	SpawnWeightValues []int  `json:"SpawnWeight_Values"`
	Wordlist          int    `json:"Wordlist"`
	Hash32            int    `json:"HASH32"`
	Key               int    `json:"_key"`
}
