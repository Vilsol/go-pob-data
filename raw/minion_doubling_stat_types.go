package raw

type MinionDoublingStatType struct {
	Stat       *int   `json:"Stat"`
	MinionType *int   `json:"MinionType"`
	Name       string `json:"Name"`
	ID         int    `json:"Id"`
	Key        int    `json:"_key"`
}
