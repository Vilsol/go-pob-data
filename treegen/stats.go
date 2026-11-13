package treegen

import (
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	st "github.com/Vilsol/go-pob-data/stat_translations"
)

const (
	// passiveDescriptions describes passive skill and mastery effect stats.
	passiveDescriptions = "Metadata/StatDescriptions/passive_skill_stat_descriptions.txt"
	// auraDescriptions describes the buffs passives apply around the player ("Nearby Enemies ...");
	// it includes passiveDescriptions and rewords some of its stats.
	auraDescriptions = "Metadata/StatDescriptions/passive_skill_aura_stat_descriptions.txt"
)

// ClientStrings rows for the line a node granting passive points shows below its stats.
const (
	grantsPassivePoint  = "PassiveNodeGrantsPassivePoint"
	grantsPassivePoints = "PassiveNodeGrantsPassivePoints"
)

// ClientStrings rows the client builds a granted skill's line from, which 3.30 added when it dropped the
// local_display_grants_skill_* stat descriptions.
const (
	grantsEffect          = "ModGrantsEffect"
	grantsEffectWithLevel = "ModGrantsEffectWithLevel"
)

// grants are the skills passive skills and mastery effects grant, nil when the game has no grantsEffect line.
type grants struct {
	text           string
	textWithLevel  string
	passives       []passiveGrant
	masteryEffects []masteryGrant
	levels         []grantedEffectLevel
	effects        []grantedEffect
	activeSkills   []activeSkill
}

func loadGrants(l *tableLoader, clientStrings []clientString) *grants {
	g := &grants{}
	for _, c := range clientStrings {
		switch c.ID {
		case grantsEffect:
			g.text = c.Text
		case grantsEffectWithLevel:
			g.textWithLevel = c.Text
		}
	}
	if g.text == "" && g.textWithLevel == "" {
		return nil
	}
	if g.text == "" || g.textWithLevel == "" {
		l.err = fmt.Errorf("ClientStrings has only one of %s and %s", grantsEffect, grantsEffectWithLevel)
		return nil
	}
	loadTable(l, "PassiveSkills", &g.passives)
	loadTable(l, "PassiveSkillMasteryEffects", &g.masteryEffects)
	loadTable(l, "GrantedEffectsPerLevel", &g.levels)
	loadTable(l, "GrantedEffects", &g.effects)
	loadTable(l, "ActiveSkills", &g.activeSkills)
	return g
}

// line is "Grants [Level N] [Id::N|Name]" for a GrantedEffectsPerLevel row; the level is left out for a skill
// that cannot be supported.
func (g *grants) line(levelRow int) (string, error) {
	lvl := g.levels[levelRow]
	e := g.effects[lvl.GrantedEffect]
	if e.ActiveSkill == nil {
		return "", fmt.Errorf("granted effect %s has no active skill", e.ID)
	}
	link := fmt.Sprintf("[%s::%d|%s]", e.ID, lvl.Level, g.activeSkills[*e.ActiveSkill].DisplayedName)
	if e.CannotBeSupported {
		return strings.ReplaceAll(g.text, "{0}", link), nil
	}
	return strings.NewReplacer("{0}", link, "{1}", strconv.Itoa(lvl.Level)).Replace(g.textWithLevel), nil
}

// grantLines is the line naming the skill a GrantedEffectsPerLevel row grants, if there is one.
func (d *gameData) grantLines(grant *int) ([]string, error) {
	if grant == nil {
		return nil, nil
	}
	line, err := d.grants.line(*grant)
	if err != nil {
		return nil, err
	}
	return []string{line}, nil
}

type descriptions struct {
	passive, aura *st.TranslationParser
}

func loadDescriptions(loader fs.FS, gameVersion string) (*descriptions, error) {
	passive, err := st.NewDescriber(loader, gamePath(gameVersion, passiveDescriptions))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load %s", passiveDescriptions)
	}
	aura, err := st.NewDescriber(loader, gamePath(gameVersion, auraDescriptions))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load %s", auraDescriptions)
	}
	return &descriptions{passive: passive, aura: aura}, nil
}

func (d *gameData) statIDs(stats []int) []string {
	ids := make([]string, len(stats))
	for i, s := range stats {
		ids[i] = d.stats[s].ID
	}
	return ids
}

// describeNode fills the node's stat lines and reminder text: the skill it grants, its own stats, then
// the buffs it grants, then the passive points it grants. Reminders follow the lines that reference them,
// then the node's ReminderStrings, each text once.
func (d *gameData) describeNode(node *Node, p *passiveSkill, grant *int) error {
	granted, err := d.grantLines(grant)
	if err != nil {
		return err
	}
	node.Stats = append(node.Stats, granted...)

	described, err := d.descriptions.passive.Describe(d.statIDs(p.Stats), p.statValues()[:len(p.Stats)])
	if err != nil {
		return errors.Wrap(err, "failed to describe stats")
	}
	node.Stats = append(node.Stats, described.Lines...)
	reminders := described.Reminders
	node.Undescribed = described.Undescribed

	for _, b := range p.PassiveSkillBuffs {
		template := d.buffTemplates[b]
		if template.BuffDefinition == nil {
			return fmt.Errorf("buff template %d has no buff definition", b)
		}
		def := d.buffDefinitions[*template.BuffDefinition]
		if len(template.StatValues) < len(def.Stats) {
			return fmt.Errorf("buff template %d has %d values for %d stats", b, len(template.StatValues), len(def.Stats))
		}
		// Granted flags are stats the buff always sets to 1.
		values := slices.Concat(template.StatValues[:len(def.Stats)], slices.Repeat([]int{1}, len(def.GrantedFlags)))
		buff, err := d.descriptions.aura.Describe(slices.Concat(d.statIDs(def.Stats), d.statIDs(def.GrantedFlags)), values)
		if err != nil {
			return errors.Wrapf(err, "failed to describe buff template %d", b)
		}
		node.Stats = append(node.Stats, buff.Lines...)
		reminders = append(reminders, buff.Reminders...)
		node.Undescribed = append(node.Undescribed, buff.Undescribed...)
	}

	if p.SkillPointsGranted > 0 {
		id := grantsPassivePoints
		if p.SkillPointsGranted == 1 {
			id = grantsPassivePoint
		}
		text, err := d.clientString(id)
		if err != nil {
			return err
		}
		node.Stats = append(node.Stats, strings.ReplaceAll(text, "{0}", strconv.Itoa(p.SkillPointsGranted)))
	}

	for _, r := range p.ReminderStrings {
		reminders = append(reminders, d.reminderTexts[r].ID)
	}
	node.ReminderText, err = d.reminderLines(reminders)
	return err
}

func (d *gameData) describeMasteryEffect(e *masteryEffect, grant *int) (*MasteryEffect, error) {
	if len(e.Stats) > len(e.statValues()) {
		return nil, fmt.Errorf("%d stats but only %d stat value columns", len(e.Stats), len(e.statValues()))
	}
	described, err := d.descriptions.passive.Describe(d.statIDs(e.Stats), e.statValues()[:len(e.Stats)])
	if err != nil {
		return nil, errors.Wrap(err, "failed to describe stats")
	}
	if len(described.Undescribed) > 0 {
		return nil, fmt.Errorf("no description for %s", strings.Join(described.Undescribed, ", "))
	}
	reminders, err := d.reminderLines(described.Reminders)
	if err != nil {
		return nil, err
	}
	granted, err := d.grantLines(grant)
	if err != nil {
		return nil, err
	}
	return &MasteryEffect{Effect: e.Hash16, Stats: slices.Concat([]string{}, granted, described.Lines), ReminderText: reminders}, nil
}

// reminderLines resolves ReminderText ids, dropping repeats, into one entry per text line.
func (d *gameData) reminderLines(ids []string) ([]string, error) {
	var lines []string
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		text, ok := d.reminderByID[id]
		if !ok {
			return nil, fmt.Errorf("no ReminderText row %s", id)
		}
		lines = append(lines, strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")...)
	}
	return lines, nil
}

func (d *gameData) clientString(id string) (string, error) {
	i := slices.IndexFunc(d.clientStrings, func(c clientString) bool { return c.ID == id })
	if i < 0 {
		return "", fmt.Errorf("no ClientStrings row %s", id)
	}
	return d.clientStrings[i].Text, nil
}
