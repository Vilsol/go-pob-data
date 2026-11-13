package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type BaseItemType struct {
	SoundEffect                   *int          `json:"SoundEffect"`
	FlavourTextKey                *int          `json:"FlavourTextKey"`
	FragmentBaseItemTypesKey      *int          `json:"FragmentBaseItemTypesKey"`
	ID                            string        `json:"Id"`
	Name                          string        `json:"Name"`
	Inflection                    string        `json:"Inflection"`
	InheritsFrom                  string        `json:"InheritsFrom"`
	TagsKeys                      []int         `json:"TagsKeys"`
	EnchantMods                   []int         `json:"TalismanEnchants"` // TalismanEnchants in the schema (3.29+)
	IdentifyMagicAchievementItems []interface{} `json:"IdentifyMagic_AchievementItems"`
	IdentifyAchievementItems      []interface{} `json:"Identify_AchievementItems"`
	ImplicitModsKeys              []int         `json:"Implicit_ModsKeys"`
	VendorRecipeAchievementItems  []int         `json:"VendorRecipe_AchievementItems"`
	EquipAchievements             []int         `json:"EquipAchievements"`
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

// UnmarshalJSON reads EquipAchievements from Equip_AchievementItemsKey too, the single key tables extracted
// before 3.30 hold.
func (b *BaseItemType) UnmarshalJSON(data []byte) error {
	type plain BaseItemType
	var named struct {
		EquipAchievementItemsKey *int `json:"Equip_AchievementItemsKey"`
		plain
	}
	if err := json.Unmarshal(data, &named); err != nil {
		return errors.Wrap(err, "failed to decode base item type")
	}
	*b = BaseItemType(named.plain)
	if named.EquipAchievementItemsKey != nil {
		b.EquipAchievements = []int{*named.EquipAchievementItemsKey}
	}
	return nil
}
