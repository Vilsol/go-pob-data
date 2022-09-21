package raw

type CraftingBenchOption struct {
	AddEnchantment              *int   `json:"AddEnchantment"`
	AddMod                      *int   `json:"AddMod"`
	UnlockCategory              *int   `json:"UnlockCategory"`
	Description                 string `json:"Description"`
	SocketColours               string `json:"SocketColours"`
	Name                        string `json:"Name"`
	CostBaseItemTypes           []int  `json:"Cost_BaseItemTypes"`
	CostValues                  []int  `json:"Cost_Values"`
	CraftingItemClassCategories []int  `json:"CraftingItemClassCategories"`
	RecipeIDS                   []int  `json:"RecipeIds"`
	ItemClasses                 []int  `json:"ItemClasses"`
	Links                       int    `json:"Links"`
	CraftingBenchCustomAction   int    `json:"CraftingBenchCustomAction"`
	Key                         int    `json:"_key"`
	Order                       int    `json:"Order"`
	UnveilsRequired2            int    `json:"UnveilsRequired2"`
	RequiredLevel               int    `json:"RequiredLevel"`
	HideoutNPCSKey              int    `json:"HideoutNPCsKey"`
	Sockets                     int    `json:"Sockets"`
	SortCategory                int    `json:"SortCategory"`
	Tier                        int    `json:"Tier"`
	ItemQuantity                int    `json:"ItemQuantity"`
	UnveilsRequired             int    `json:"UnveilsRequired"`
	IsAreaOption                bool   `json:"IsAreaOption"`
	IsDisabled                  bool   `json:"IsDisabled"`
}
