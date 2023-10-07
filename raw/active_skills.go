package raw

type ActiveSkill struct {
	AlternateSkillTargetingBehavioursKey *int   `json:"AlternateSkillTargetingBehavioursKey"`
	AIFile                               string `json:"AIFile"`
	WebsiteImage                         string `json:"WebsiteImage"`
	Description                          string `json:"Description"`
	DisplayedName                        string `json:"DisplayedName"`
	IconDDSFile                          string `json:"Icon_DDSFile"`
	ID                                   string `json:"Id"`
	WebsiteDescription                   string `json:"WebsiteDescription"`
	SkillID                              string `json:"Var3"`
	WeaponRestrictionItemClassesKeys     []int  `json:"WeaponRestriction_ItemClassesKeys"`
	MinionActiveSkillTypes               []int  `json:"MinionActiveSkillTypes"`
	OutputStatKeys                       []int  `json:"Output_StatKeys"`
	InputStatKeys                        []int  `json:"Input_StatKeys"`
	ActiveSkillTypes                     []int  `json:"ActiveSkillTypes"`
	ActiveSkillTargetTypes               []int  `json:"ActiveSkillTargetTypes"`
	SkillTotemID                         int    `json:"SkillTotemId"`
	Key                                  int    `json:"_key"`
	IsManuallyCasted                     bool   `json:"IsManuallyCasted"`
}
