package raw

type BaseItemType struct {
	SoundEffect                   *int          `json:"SoundEffect"`
	EquipAchievementItemsKey      *int          `json:"Equip_AchievementItemsKey"`
	FlavourTextKey                *int          `json:"FlavourTextKey"`
	FragmentBaseItemTypesKey      *int          `json:"FragmentBaseItemTypesKey"`
	ID                            string        `json:"Id"`
	Name                          string        `json:"Name"`
	Inflection                    string        `json:"Inflection"`
	InheritsFrom                  string        `json:"InheritsFrom"`
	TagsKeys                      []int         `json:"TagsKeys"`
	IdentifyMagicAchievementItems []interface{} `json:"IdentifyMagic_AchievementItems"`
	IdentifyAchievementItems      []interface{} `json:"Identify_AchievementItems"`
	ImplicitModsKeys              []int         `json:"Implicit_ModsKeys"`
	VendorRecipeAchievementItems  []int         `json:"VendorRecipe_AchievementItems"`
	SizeOnGround                  int           `json:"SizeOnGround"`
	ItemVisualIdentity            int           `json:"ItemVisualIdentity"`
	ModDomain                     int           `json:"ModDomain"`
	ItemClassesKey                int           `json:"ItemClassesKey"`
	SiteVisibility                int           `json:"SiteVisibility"`
	DropLevel                     int           `json:"DropLevel"`
	Height                        int           `json:"Height"`
	Hash                          int           `json:"HASH"`
	Width                         int           `json:"Width"`
	Key                           int           `json:"_key"`
	IsCorrupted                   bool          `json:"IsCorrupted"`
}
