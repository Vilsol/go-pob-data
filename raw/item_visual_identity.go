package raw

// ItemVisualIdentity keeps only the columns that identify an item's art; the rest are client
// rendering details (models, sounds, achievements).
type ItemVisualIdentity struct {
	ID             string `json:"Id"`
	DDSFile        string `json:"DDSFile"`
	Hash16         int    `json:"HASH16"`
	Key            int    `json:"_key"`
	IsAlternateArt bool   `json:"IsAlternateArt"`
}
