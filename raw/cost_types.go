package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type CostType struct {
	FormatText string `json:"FormatText"`
	ID         string `json:"Id"`
	StatsKey   int    `json:"StatsKey"`
	Divisor    int    `json:"Var3"`
	Key        int    `json:"_key"`
}

// UnmarshalJSON reads Divisor under its schema name too: tables extracted before dat-schema named the
// column call it Var3.
func (c *CostType) UnmarshalJSON(b []byte) error {
	type plain CostType
	var named struct {
		Divisor *int `json:"Divisor"`
		plain
	}
	if err := json.Unmarshal(b, &named); err != nil {
		return errors.Wrap(err, "failed to decode cost type")
	}
	*c = CostType(named.plain)
	if named.Divisor != nil {
		c.Divisor = *named.Divisor
	}
	return nil
}
