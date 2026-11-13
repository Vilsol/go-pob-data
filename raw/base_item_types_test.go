package raw

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestBaseItemTypeEquipAchievementLayouts(t *testing.T) {
	for _, b := range []string{
		`{"Id":"Ring1","Equip_AchievementItemsKey":7320}`,
		`{"Id":"Ring1","EquipAchievements":[7320]}`,
	} {
		var i BaseItemType
		if err := json.Unmarshal([]byte(b), &i); err != nil {
			t.Fatal(err)
		}
		if i.ID != "Ring1" || !slices.Equal(i.EquipAchievements, []int{7320}) {
			t.Errorf("%s decoded to %+v", b, i)
		}
	}
}
