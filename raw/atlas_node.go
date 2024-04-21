package raw

type AtlasNode struct {
	WorldAreasKey         int     `json:"WorldAreasKey"`
	ItemVisualIdentityKey int     `json:"ItemVisualIdentityKey"`
	Var2                  bool    `json:"Var2"`
	MapsKey               int     `json:"MapsKey"`
	FlavourTextKey        int     `json:"FlavourTextKey"`
	AtlasNodeKeys         []int   `json:"AtlasNodeKeys"`
	Tier0                 int     `json:"Tier0"`
	Tier1                 int     `json:"Tier1"`
	Tier2                 int     `json:"Tier2"`
	Tier3                 int     `json:"Tier3"`
	Tier4                 int     `json:"Tier4"`
	Var11                 float64 `json:"Var11"`
	Var12                 float64 `json:"Var12"`
	Var13                 float64 `json:"Var13"`
	Var14                 float64 `json:"Var14"`
	Var15                 float64 `json:"Var15"`
	DDSFile               *string `json:"DDSFile"`
	Var17                 bool    `json:"Var17"`
	NotOnAtlas            bool    `json:"NotOnAtlas"`
	Var19                 int     `json:"Var19"`
	Var20                 int     `json:"Var20"`
	Var21                 int     `json:"Var21"`
	Var22                 int     `json:"Var22"`
	Key                   int     `json:"_key"`
}
