package st

import (
	"testing"
	"testing/fstest"

	"github.com/MarvinJWendt/testza"
	"golang.org/x/text/encoding/unicode"
)

const baseDescriptions = `no_description hidden_stat
description
	1 damage_+%
	2
		1|# "{0}% increased Damage"
		#|-1 "{0}% reduced Damage" negate 1
description
	1 regen_per_minute
	1
		# "Regenerate {0} Life per second" per_minute_to_per_second 1 reminderstring ReminderTextA
description
	2 min_added max_added
	2
		0 0 "never shown"
		# # "Adds {0} to {1} Damage"
description
	1 overridden
	1
		# "base wording"
description
	1 range_+
	1
		# "{0:+d} metres to Range" divide_by_ten_1dp_if_required 1
description
	1 max_stacks
	1
		# "up to {0} stacks" add_one 1
`

const overrideDescriptions = `include "Metadata/StatDescriptions/base.txt"
description
	1 overridden
	1
		# "{0:+d} override wording" canonical_line reminderstring ReminderTextB
	lang "French"
	1
		# "{0:+d} French wording"
description
	1 duration_ms
	2
		!1000 "Lasts {0} seconds" milliseconds_to_seconds_2dp 1
		# "Lasts {0} second" milliseconds_to_seconds_if_required 1
`

func utf16(t *testing.T, s string) []byte {
	t.Helper()
	b, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder().Bytes([]byte(s))
	testza.AssertNoError(t, err)
	return b
}

func TestDescribe(t *testing.T) {
	loader := fstest.MapFS{
		"metadata/statdescriptions/base.txt":     {Data: utf16(t, baseDescriptions)},
		"metadata/statdescriptions/override.txt": {Data: utf16(t, overrideDescriptions)},
	}
	describer, err := NewDescriber(loader, "metadata/statdescriptions/override.txt")
	testza.AssertNoError(t, err)

	tests := []struct {
		name   string
		ids    []string
		values []int
		want   Described
	}{
		{
			name:   "negate picks the matching variant",
			ids:    []string{"damage_+%"},
			values: []int{-10},
			want:   Described{Lines: []string{"10% reduced Damage"}},
		},
		{
			name:   "overrides sort after the included file and keep their reminders",
			ids:    []string{"overridden", "regen_per_minute", "damage_+%"},
			values: []int{3, 90, 5},
			want: Described{
				Lines:     []string{"5% increased Damage", "Regenerate 1.5 Life per second", "+3 override wording"},
				Reminders: []string{"ReminderTextA", "ReminderTextB"},
			},
		},
		{
			name:   "a multi-stat description renders once, missing stats count as zero",
			ids:    []string{"max_added", "min_added"},
			values: []int{12, 4},
			want:   Described{Lines: []string{"Adds 4 to 12 Damage"}},
		},
		{
			name:   "zero, hidden and undescribed stats show nothing",
			ids:    []string{"damage_+%", "hidden_stat", "unknown_stat"},
			values: []int{0, 1, 1},
			want:   Described{Undescribed: []string{"unknown_stat"}},
		},
		{
			name:   "tenths and an added one",
			ids:    []string{"range_+", "max_stacks"},
			values: []int{4, 1},
			want:   Described{Lines: []string{"+0.4 metres to Range", "up to 2 stacks"}},
		},
		{
			name:   "negated limits and fixed decimals",
			ids:    []string{"duration_ms"},
			values: []int{2500},
			want:   Described{Lines: []string{"Lasts 2.50 seconds"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := describer.Describe(tt.ids, tt.values)
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, tt.want, *got)
		})
	}

	_, err = describer.Describe([]string{"duration_ms"}, []int{1000})
	testza.AssertNotNil(t, err, "an unsupported handler on the chosen variant must fail")
}
