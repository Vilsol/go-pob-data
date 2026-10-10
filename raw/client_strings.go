package raw

type ClientString struct {
	ID              string `json:"Id"`
	Text            string `json:"Text"`
	XBoxText        string `json:"XBoxText"`
	XBoxText2       string `json:"XBoxText2"`
	PlaystationText string `json:"PlaystationText"`
	Hash32          int    `json:"HASH32"`
	Key             int    `json:"_key"`
}
