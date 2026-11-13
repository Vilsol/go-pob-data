package st

import (
	"testing"
	"testing/fstest"

	"github.com/MarvinJWendt/testza"

	"github.com/Vilsol/go-pob-data/raw"
	"github.com/Vilsol/go-pob-data/utils"
)

const skillDescriptions = `has_identifiers

no_description hidden_stat
description damage_and_speed
	2 damage_+% speed_+%
	2
		1|99 # gem_quality "{0}% more Damage" negate 1 negate 2
		#|-1 !0 "{0}% less Damage" canonical_line reminderstring ReminderTextA
	lang "French"
	1
		# # "{0}% de dégâts"
handed_description weapon_damage_range
	2
main_min main_max
		off_min off_max
	1
		# # "{0} to {1} Damage"
`

func parse(t *testing.T, text string) (*TranslationParser, error) {
	t.Helper()
	loader := fstest.MapFS{"metadata/statdescriptions/skill.txt": {Data: utf16(t, text)}}
	p := NewTranslationParser(loader)
	return p, p.ParseFile("metadata/statdescriptions/skill.txt")
}

func TestParseFile(t *testing.T) {
	p, err := parse(t, skillDescriptions)
	testza.AssertNoError(t, err)
	testza.AssertLen(t, p.ordered, 2)

	d := p.ordered[0]
	testza.AssertEqual(t, "damage_and_speed", d.Identifier)
	testza.AssertEqual(t, []string{"damage_+%", "speed_+%"}, d.Stats)
	testza.AssertEqual(t, []raw.LangTranslation{
		{
			String:        "{0}% more Damage",
			Qualifier:     "gem_quality",
			Conditions:    []raw.Condition{{Min: utils.Ptr(1), Max: utils.Ptr(99)}, {}},
			Handlers:      []raw.IndexHandler{{Name: "negate", Arg: "1"}, {Name: "negate", Arg: "2"}},
			IndexHandlers: raw.IndexHandlers{"negate": "2"},
		},
		{
			String:        "{0}% less Damage",
			Conditions:    []raw.Condition{{Max: utils.Ptr(-1)}, {Min: utils.Ptr(0), Max: utils.Ptr(0), Negated: true}},
			Handlers:      []raw.IndexHandler{{Name: "canonical_line"}, {Name: "reminderstring", Arg: "ReminderTextA"}},
			IndexHandlers: raw.IndexHandlers{"canonical_line": "", "reminderstring": "ReminderTextA"},
		},
	}, d.Lang["English"])
	testza.AssertLen(t, d.Lang["French"], 1)

	handed := p.ordered[1]
	testza.AssertEqual(t, []string{"main_min", "main_max"}, handed.Stats)
	testza.AssertEqual(t, []string{"off_min", "off_max"}, handed.OffHand)
	testza.AssertEqual(t, handed, p.descriptors["main_max"])
	testza.AssertNil(t, p.descriptors["hidden_stat"].Stats)
	testza.AssertEqual(t, []string{"hidden_stat"}, p.noDescription)
}

func TestParseFileRejects(t *testing.T) {
	for name, text := range map[string]string{
		"unknown directive":    "table_flags 1\n",
		"too few variants":     "description\n\t1 a\n\t2\n\t\t# \"x\"\n",
		"limits do not match":  "description\n\t2 a b\n\t1\n\t\t# \"x\"\n",
		"stat count mismatch":  "description\n\t2 a\n\t1\n\t\t# \"x\"\n",
		"off hand mismatch":    "handed_description\n\t2 a b\n\t\tc\n\t1\n\t\t# # \"x\"\n",
		"bad limit":            "description\n\t1 a\n\t1\n\t\t1|x \"x\"\n",
		"line outside a block": "description\n\t1 a\n\t1\n\t\t# \"x\"\n\t\t# \"y\"\n",
		"repeat differs":       "description\n\t1 a\n\t1\n\t\t# \"x\"\n\tlang \"German\"\n\t1\n\t\t# \"y\"\n\tlang \"German\"\n\t1\n\t\t# \"z\"\n",
		"repeat is shorter":    "description\n\t1 a\n\t1\n\t\t# \"x\"\n\tlang \"German\"\n\t1\n\t\t# \"y\"\n\tlang \"German\"\n\t0\n",
		"english repeated":     "description\n\t1 a\n\t1\n\t\t# \"x\"\n\tlang \"English\"\n\t1\n\t\t# \"x\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parse(t, text)
			testza.AssertNotNil(t, err)
		})
	}
}

// PoE2's atlas_variant_stat_descriptions.csd repeats a language block with the same lines.
func TestParseFileRepeatedLanguage(t *testing.T) {
	p, err := parse(t, "description\n\t1 a\n\t1\n\t\t# \"x\"\n\tlang \"German\"\n\t1\n\t\t# \"y\"\n\tlang \"French\"\n\t1\n\t\t# \"z\"\n\tlang \"German\"\n\t1\n\t\t# \"y\"\n")
	testza.AssertNoError(t, err)
	testza.AssertLen(t, p.ordered[0].Lang["German"], 1)
	testza.AssertEqual(t, "y", p.ordered[0].Lang["German"][0].String)
}

// PoE2's .csd files include each other under Data/StatDescriptions.
func TestParseFileCsd(t *testing.T) {
	text := "include \"Data/StatDescriptions/stat_descriptions.csd\"\ndescription\n\t1 a\n\t1\n\t\t# \"x\"\n"
	loader := fstest.MapFS{"data/statdescriptions/skill.csd": {Data: utf16(t, text)}}
	p := NewTranslationParser(loader)
	testza.AssertNoError(t, p.ParseFile("data/statdescriptions/skill.csd"))
	testza.AssertEqual(t, []string{"Data/StatDescriptions/stat_descriptions.csd"}, p.includes)
	testza.AssertLen(t, p.ordered, 1)

	_, err := parse(t, "include \"Data/StatDescriptions/stat_descriptions.txt\"\n")
	testza.AssertNotNil(t, err)
}

// 3.30's PoE1 files include each other as Metadata/StatDescriptions/*.csd.
func TestParseFileIncludesCsd(t *testing.T) {
	text := "include \"Metadata/StatDescriptions/stat_descriptions.csd\"\ndescription\n\t1 a\n\t1\n\t\t# \"x\"\n"
	p, err := parse(t, text)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{"Metadata/StatDescriptions/stat_descriptions.csd"}, p.includes)
	testza.AssertLen(t, p.ordered, 1)
}

// An include or a listed file resolves to the same name with the other of .txt and .csd when only that
// one exists, under Metadata/StatDescriptions only.
func TestParseFileOtherExtension(t *testing.T) {
	base := "description\n\t1 a\n\t1\n\t\t# \"base {0}\"\n"
	child := "include \"Metadata/StatDescriptions/base.csd\"\ndescription\n\t1 b\n\t1\n\t\t# \"child {0}\"\n"
	loader := fstest.MapFS{
		"metadata/statdescriptions/base.txt":  {Data: utf16(t, base)},
		"metadata/statdescriptions/child.csd": {Data: utf16(t, child)},
		"data/statdescriptions/only.txt":      {Data: utf16(t, base)},
	}
	d, err := NewDescriber(loader, "metadata/statdescriptions/child.txt")
	testza.AssertNoError(t, err)
	got, err := d.Describe([]string{"a", "b"}, []int{1, 2})
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{"base 1", "child 2"}, got.Lines)

	testza.AssertNotNil(t, NewTranslationParser(loader).ParseFile("metadata/statdescriptions/missing.txt"))
	testza.AssertNotNil(t, NewTranslationParser(loader).ParseFile("data/statdescriptions/only.csd"))
}

// 3.30's stat_descriptions.txt writes `1|99#` for `1|99`.
func TestParseFileStrayHashAfterMax(t *testing.T) {
	p, err := parse(t, "description\n\t1 a\n\t2\n\t\t1|99# \"{0}% chance\"\n\t\t100|# \"always\"\n")
	testza.AssertNoError(t, err)
	c := p.ordered[0].Lang["English"][0].Conditions[0]
	testza.AssertEqual(t, 1, *c.Min)
	testza.AssertEqual(t, 99, *c.Max)
	testza.AssertEqual(t, 99, p.ordered[0].variants[0].limits[0].max)

	_, err = parse(t, "description\n\t1 a\n\t1\n\t\t1#|99 \"x\"\n")
	testza.AssertNotNil(t, err)
}
