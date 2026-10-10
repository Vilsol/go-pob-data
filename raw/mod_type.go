package raw

type ModType struct {
	Name                  string `json:"Name"`
	Stats                 []int  `json:"Stats"`
	ModSellPriceTypesKeys []int  `json:"ModSellPriceTypesKeys"`
	Key                   int    `json:"_key"`
}
