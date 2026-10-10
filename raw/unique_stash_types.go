package raw

type UniqueStashType struct {
	ID                   string `json:"Id"`
	Name                 string `json:"Name"`
	Image                string `json:"Image"`
	Order                int    `json:"Order"`
	Width                int    `json:"Width"`
	Height               int    `json:"Height"`
	TotalCount           int    `json:"TotalCount"`
	StandardCount        int    `json:"StandardCount"`
	ChallengeLeagueCount int    `json:"ChallengeLeagueCount"`
	Key                  int    `json:"_key"`
	IsDisabled           bool   `json:"IsDisabled"`
}
