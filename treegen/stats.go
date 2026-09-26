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

// describeNode fills the node's stat lines and reminder text: its own stats, then the buffs it
// grants, then the passive points it grants. Reminders follow the lines that reference them,
// then the node's ReminderStrings, each text once.
func (d *gameData) describeNode(node *Node, p *passiveSkill) error {
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

func (d *gameData) describeMasteryEffect(e *masteryEffect) (*MasteryEffect, error) {
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
	return &MasteryEffect{Effect: e.Hash16, Stats: append([]string{}, described.Lines...), ReminderText: reminders}, nil
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
