package raw

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Stat struct {
	MainHandAliasStatsKey *int     `json:"MainHandAlias_StatsKey"`
	Category              *int     `json:"Category"`
	OffHandAliasStatsKey  *int     `json:"OffHandAlias_StatsKey"`
	ID                    string   `json:"Id"`
	Text                  string   `json:"Text"`
	ContextFlags          []int    `json:"ContextFlags"`
	BelongsStatsKey       []string `json:"BelongsStatsKey"`
	Hash32                int      `json:"HASH32"`
	Semantics             int      `json:"Semantics"`
	Key                   int      `json:"_key"`
	IsWeaponLocal         bool     `json:"IsWeaponLocal"`
	IsVirtual             bool     `json:"IsVirtual"`
	IsScalable            bool     `json:"IsScalable"`
	IsLocal               bool     `json:"IsLocal"`
	CannotGrantToMinion   bool     `json:"CannotGrantToMinion"`
}

// UnmarshalJSON reads CannotGrantToMinion from the unnamed column the schema of each layout leaves it
// in: Var13 from 3.27, Var14 in 3.21-3.26 (the layouts with Var10) and Var15 in 3.18-3.20 (with Var11).
func (s *Stat) UnmarshalJSON(b []byte) error {
	type plain Stat
	if err := json.Unmarshal(b, (*plain)(s)); err != nil {
		return errors.Wrap(err, "failed to decode stat")
	}

	var columns map[string]json.RawMessage
	if err := json.Unmarshal(b, &columns); err != nil {
		return errors.Wrap(err, "failed to decode stat")
	}
	column := ""
	switch {
	case columns["Var13"] != nil:
		column = "Var13"
	case columns["Var10"] != nil:
		column = "Var14"
	case columns["Var11"] != nil:
		column = "Var15"
	}
	if raw, ok := columns[column]; ok {
		return errors.Wrap(json.Unmarshal(raw, &s.CannotGrantToMinion), "failed to decode CannotGrantToMinion")
	}
	return nil
}
