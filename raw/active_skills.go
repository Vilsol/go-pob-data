package raw

type ActiveSkill struct {
	AIFile                               string `json:"AIFile"`
	ActiveSkillTargetTypes               []int  `json:"ActiveSkillTargetTypes"`
	ActiveSkillTypes                     []int  `json:"ActiveSkillTypes"`
	AlternateSkillTargetingBehavioursKey *int   `json:"AlternateSkillTargetingBehavioursKey"`
	Description                          string `json:"Description"`
	DisplayedName                        string `json:"DisplayedName"`
	IconDDSFile                          string `json:"Icon_DDSFile"`
	ID                                   string `json:"Id"`
	InputStatKeys                        []int  `json:"Input_StatKeys"`
	IsManuallyCasted                     bool   `json:"IsManuallyCasted"`
	MinionActiveSkillTypes               []int  `json:"MinionActiveSkillTypes"`
	OutputStatKeys                       []int  `json:"Output_StatKeys"`
	SkillTotemID                         int    `json:"SkillTotemId"`
	WeaponRestrictionItemClassesKeys     []int  `json:"WeaponRestriction_ItemClassesKeys"`
	WebsiteDescription                   string `json:"WebsiteDescription"`
	WebsiteImage                         string `json:"WebsiteImage"`
	Key                                  int    `json:"_key"`
}
