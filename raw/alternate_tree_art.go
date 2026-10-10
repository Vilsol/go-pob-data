package raw

type AlternateTreeArt struct {
	ConnectionArt      *int `json:"ConnectionArt"`
	KeystoneFrameArt   *int `json:"KeystoneFrameArt"`
	NotableFrameArt    *int `json:"NotableFrameArt"`
	AscendancyFrameArt *int `json:"AscendancyFrameArt"`
	RadiiArt           *int `json:"RadiiArt"`
	TreeVersion        int  `json:"TreeVersion"`
	Key                int  `json:"_key"`
}
