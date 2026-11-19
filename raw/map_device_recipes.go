package raw

type MapDeviceRecipe struct {
	WorldArea   *int   `json:"WorldArea"`
	AtlasNode   *int   `json:"AtlasNode"`
	ID          string `json:"Id"`
	RecipeItems []int  `json:"RecipeItems"`
	Mods        []int  `json:"Mods"`
	AreaLevel   int    `json:"AreaLevel"`
	Key         int    `json:"_key"`
}
