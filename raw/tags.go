package raw

type Tag struct {
	DisplayString string  `json:"DisplayString"`
	ID            string  `json:"Id"`
	Name          TagName `json:"Name"`
	Key           int     `json:"_key"`
}

type TagName string

const (
	TagAilment    TagName = "Ailment"
	TagAttack     TagName = "Attack"
	TagAttribute  TagName = "Attribute"
	TagAura       TagName = "Aura"
	TagCaster     TagName = "Caster"
	TagChaos      TagName = "Chaos"
	TagCold       TagName = "Cold"
	TagCritical   TagName = "Critical"
	TagCurse      TagName = "Curse"
	TagDamage     TagName = "Damage"
	TagDefences   TagName = "Defences"
	TagElemental  TagName = "Elemental"
	TagFire       TagName = "Fire"
	TagGem        TagName = "Gem"
	TagLife       TagName = "Life"
	TagLightning  TagName = "Lightning"
	TagMana       TagName = "Mana"
	TagMinion     TagName = "Minion"
	TagPhysical   TagName = "Physical"
	TagResistance TagName = "Resistance"
	TagSpeed      TagName = "Speed"
)
