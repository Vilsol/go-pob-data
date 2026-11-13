package raw

import (
	"encoding/json"
	"testing"
)

func TestAtlasNodeLayouts(t *testing.T) {
	for _, b := range []string{
		`{"Id":"MapA","Var32":1,"Var33":2,"Var34":3,"Var35":4}`,
		`{"Id":"MapA","MapNodeHeader":null,"Var33":true,"Var34":3,"Var35":false}`,
	} {
		var a AtlasNode
		if err := json.Unmarshal([]byte(b), &a); err != nil {
			t.Fatalf("%s: %v", b, err)
		}
		if a.ID != "MapA" || a.Var34 != 3 {
			t.Errorf("%s decoded to %+v", b, a)
		}
	}
}
