package raw

type CraftingItemClassCategory struct {
	ID                 string `json:"Id"`
	CategoryCode       string `json:"CategoryCode"`
	Text               string `json:"Text"`
	ItemClasses        []int  `json:"ItemClasses"`
	CraftingBenchTypes []int  `json:"CraftingBenchTypes"`
	Key                int    `json:"_key"`
}
