package fields

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmbeddedLists(t *testing.T) {
	if len(Versions()) == 0 {
		t.Fatal("no embedded lists")
	}

	for _, version := range Versions() {
		l, _ := For(version)
		for asset := range l.Assets {
			variant, ok := MinVariant(version, asset)
			if !ok || !strings.HasPrefix(variant, "min-") || len(variant) != len("min-")+8 {
				t.Errorf("%s %s: variant %q", version, asset, variant)
			}
		}
	}

	if _, ok := MinVariant("3.29", "raw/NoSuchTable"); ok {
		t.Error("a variant for an asset the list does not have")
	}
	if _, ok := MinVariant("0.1", "raw/Mods"); ok {
		t.Error("a variant for a version without a list")
	}
}

// plant returns the embedded 3.29 list with edit applied to its raw json.
func plant(t *testing.T, edit func(list map[string]any)) []byte {
	t.Helper()

	b, err := embedded.ReadFile("3.29.json")
	if err != nil {
		t.Fatal(err)
	}
	var list map[string]any
	if err := json.Unmarshal(b, &list); err != nil {
		t.Fatal(err)
	}
	edit(list)
	if b, err = json.Marshal(list); err != nil {
		t.Fatal(err)
	}
	return b
}

func weaponTypes(list map[string]any) map[string]any {
	return list["assets"].(map[string]any)["raw/WeaponTypes"].(map[string]any)
}

func TestParseRejects(t *testing.T) {
	tests := []struct {
		name string
		edit func(list map[string]any)
		want string
	}{
		{
			name: "planted key path",
			edit: func(list map[string]any) {
				weaponTypes(list)["keep"] = append(weaponTypes(list)["keep"].([]any), "NoSuchField")
			},
			want: `3.29 raw/WeaponTypes: key path "NoSuchField": WeaponType has no field NoSuchField`,
		},
		{
			name: "planted nested key path",
			edit: func(list map[string]any) {
				weaponTypes(list)["keep"] = []any{"Critical.Value"}
			},
			want: `3.29 raw/WeaponTypes: key path "Critical.Value": int has no fields, cannot select Value`,
		},
		{
			name: "descends into a map",
			edit: func(list map[string]any) {
				list["assets"].(map[string]any)["stat_translations/en/stat_descriptions"].(map[string]any)["keep"] = []any{"Descriptors.List.IndexHandlers.x"}
			},
			want: `3.29 stat_translations/en/stat_descriptions: key path "Descriptors.List.IndexHandlers.x": raw.IndexHandlers has no fields, cannot select x`,
		},
		{
			name: "unknown format",
			edit: func(list map[string]any) { list["format"] = "go-pob-data-fields/2" },
			want: `list format "go-pob-data-fields/2", this package reads "go-pob-data-fields/1"`,
		},
		{
			name: "unknown asset",
			edit: func(list map[string]any) {
				list["assets"].(map[string]any)["raw/NoSuchTable"] = map[string]any{"type": "Mod", "keep": []any{}}
			},
			want: `3.29 raw/NoSuchTable: unknown asset`,
		},
		{
			name: "unknown type",
			edit: func(list map[string]any) { weaponTypes(list)["type"] = "Weapon" },
			want: `3.29 raw/WeaponTypes: type "Weapon", the loader decodes it as WeaponType`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(plant(t, test.edit))
			if err == nil || err.Error() != test.want {
				t.Errorf("got %v\nwant %s", err, test.want)
			}
		})
	}
}
