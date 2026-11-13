package raw

import (
	"encoding/json"
	"testing"
)

func TestAtlasNodeLayouts(t *testing.T) {
	for _, b := range []string{
		`{"Id":"MapA","Var32":11,"Var33":0,"Var34":0,"Var35":0,"Var36":true,"Var37":1911}`,
		`{"Id":"MapA","MapNodeHeader":11,"Var33":true,"Var34":1990,"Var35":false}`,
	} {
		var a AtlasNode
		if err := json.Unmarshal([]byte(b), &a); err != nil {
			t.Fatalf("%s: %v", b, err)
		}
		if a.ID != "MapA" || a.Var32 != 11 || !a.Var36 || a.Var33 != 0 || a.Var34 != 0 || a.Var35 != 0 {
			t.Errorf("%s decoded to %+v", b, a)
		}
	}
}
