package raw

import (
	"encoding/json"
	"testing"
)

func TestCostTypeDivisorNames(t *testing.T) {
	for _, b := range []string{`{"Id":"Mana","Var3":4}`, `{"Id":"Mana","Divisor":4}`} {
		var c CostType
		if err := json.Unmarshal([]byte(b), &c); err != nil {
			t.Fatal(err)
		}
		if c.ID != "Mana" || c.Divisor != 4 {
			t.Errorf("%s decoded to %+v", b, c)
		}
	}
}
