package raw

type BaseItemType struct {
	DropLevel                     int           `json:"DropLevel"`
	EquipAchievementItemsKey      *int          `json:"Equip_AchievementItemsKey"`
	FlavourTextKey                *int          `json:"FlavourTextKey"`
	FragmentBaseItemTypesKey      *int          `json:"FragmentBaseItemTypesKey"`
	Hash                          int           `json:"HASH"`
	Height                        int           `json:"Height"`
	ID                            string        `json:"Id"`
	IdentifyMagicAchievementItems []interface{} `json:"IdentifyMagic_AchievementItems"`
	IdentifyAchievementItems      []interface{} `json:"Identify_AchievementItems"`
	ImplicitModsKeys              []int         `json:"Implicit_ModsKeys"`
	Inflection                    string        `json:"Inflection"`
	InheritsFrom                  string        `json:"InheritsFrom"`
	IsCorrupted                   bool          `json:"IsCorrupted"`
	ItemClassesKey                int           `json:"ItemClassesKey"`
	ItemVisualIdentity            int           `json:"ItemVisualIdentity"`
	ModDomain                     int           `json:"ModDomain"`
	Name                          string        `json:"Name"`
	SiteVisibility                int           `json:"SiteVisibility"`
	SizeOnGround                  int           `json:"SizeOnGround"`
	SoundEffect                   *int          `json:"SoundEffect"`
	TagsKeys                      []int         `json:"TagsKeys"`
	VendorRecipeAchievementItems  []int         `json:"VendorRecipe_AchievementItems"`
	Width                         int           `json:"Width"`
	Key                           int           `json:"_key"`
}
