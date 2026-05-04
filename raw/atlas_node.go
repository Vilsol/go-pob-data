package raw

type AtlasNode struct {
	ID                   string  `json:"Id"`
	UniqueArt            string  `json:"UniqueArt"`
	NodeDDSFile          *string `json:"Node_DDSFile"`
	Connections          []int   `json:"Connections"`
	DivCards             []int   `json:"DivCards"`
	DivCardsHardmode     []int   `json:"DivCardsHardmode"`
	QuestStates          []int   `json:"QuestStates"`
	Area1                int     `json:"Area1"`
	FlavourText          int     `json:"FlavourText"`
	Tier                 int     `json:"Tier"`
	HASH16               int     `json:"HASH16"`
	Region               int     `json:"Region"`
	VoidstoneSlot        int     `json:"VoidstoneSlot"`
	MapDeviceLayout      int     `json:"MapDeviceLayout"`
	NodeDisplayName      int     `json:"NodeDisplayName"`
	Area2                int     `json:"Area2"`
	Var32                int     `json:"Var32"`
	Var33                int     `json:"Var33"`
	Var34                int     `json:"Var34"`
	Var35                int     `json:"Var35"`
	Var5                 float64 `json:"Var5"`
	Var6                 float64 `json:"Var6"`
	Var7                 float64 `json:"Var7"`
	Var8                 float64 `json:"Var8"`
	Var9                 float64 `json:"Var9"`
	Var30                float64 `json:"Var30"`
	Key                  int     `json:"_key"`
	Var2                 bool    `json:"Var2"`
	StartingNode         bool    `json:"StartingNode"`
	NotOnAtlas           bool    `json:"NotOnAtlas"`
	IsUniqueMap          bool    `json:"IsUniqueMap"`
	IsNormalMap          bool    `json:"IsNormalMap"`
	RequiresSpecificItem bool    `json:"RequiresSpecificItem"`
	Var25                bool    `json:"Var25"`
	Var28                bool    `json:"Var28"`
	Var29                bool    `json:"Var29"`
	Var31                bool    `json:"Var31"`
	Var36                bool    `json:"Var36"`
}
