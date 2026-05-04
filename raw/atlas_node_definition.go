package raw

type AtlasNodeDefinition struct {
	ID                   string  `json:"Id"`
	NodeDDSFile          *string `json:"Node_DDSFile"`
	QuestStates          []int   `json:"QuestStates"`
	Area1                int     `json:"Area1"`
	Var2                 int     `json:"Var2"`
	Var3                 int     `json:"Var3"`
	VoidstoneSlot        int     `json:"VoidstoneSlot"`
	MapDeviceLayout      int     `json:"MapDeviceLayout"`
	NodeDisplayName      int     `json:"NodeDisplayName"`
	Area2                int     `json:"Area2"`
	Header               int     `json:"Header"`
	Key                  int     `json:"_key"`
	NotOnAtlas           bool    `json:"NotOnAtlas"`
	IsUniqueMap          bool    `json:"IsUniqueMap"`
	IsNormalMap          bool    `json:"IsNormalMap"`
	RequiresSpecificItem bool    `json:"RequiresSpecificItem"`
	Var12                bool    `json:"Var12"`
	Var15                bool    `json:"Var15"`
}
