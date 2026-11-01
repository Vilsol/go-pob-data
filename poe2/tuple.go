package poe2

import "reflect"

// Every row is encoded as a strict tuple, as raw's are, and published under its loader.LayoutHash.
// TestTupleCoversAllStructs keeps these lists complete.

//msgp:tuple ActiveSkill ActiveSkillType ActiveSkillWeaponRequirement AlternatePassiveAddition AlternatePassiveSkill ArmourType AttributeRequirement BaseItemType BuffDefinition CharacterMeleeSkill ClientString ComponentCharge CostType EffectivenessCostConstant EndgameMap Flask FlavourText GameConstant GemEffect GemTag GrantedEffect GrantedEffectLabel GrantedEffectPerLevel GrantedEffectQualityStat GrantedEffectStatSet GrantedEffectStatSetPerLevel ItemClass ItemClassCategory ItemExperiencePerLevel ItemExperienceType ItemInherentSkill ItemSpirit Mod MonsterCategory MonsterPack MonsterPackEntry MonsterResistance MonsterType MonsterVariety PlayerMinionIntrinsicStat QuestStaticReward ShieldType SkillGem SkillGemSupport SoulCore SoulCoreLimit SoulCoreStat SoulCoreStatCategory SoulCoreType Stat SupportGem SupportGemFamily Tag UtilityFlaskBuff WeaponType WieldableClass WorldArea

//msgp:ignore Minion MinionModStat MinionTables SkillTables

// Tables are the row type of each table this package declares rows for, the tables whose tuple assets
// are published.
var Tables = map[string]reflect.Type{
	ActiveSkillTypeTable:               reflect.TypeFor[ActiveSkillType](),
	ActiveSkillsTable:                  reflect.TypeFor[ActiveSkill](),
	ActiveSkillWeaponRequirementTable:  reflect.TypeFor[ActiveSkillWeaponRequirement](),
	AlternatePassiveAdditionsTable:     reflect.TypeFor[AlternatePassiveAddition](),
	AlternatePassiveSkillsTable:        reflect.TypeFor[AlternatePassiveSkill](),
	ArmourTypesTable:                   reflect.TypeFor[ArmourType](),
	AttributeRequirementsTable:         reflect.TypeFor[AttributeRequirement](),
	BaseItemTypesTable:                 reflect.TypeFor[BaseItemType](),
	BuffDefinitionsTable:               reflect.TypeFor[BuffDefinition](),
	CharacterMeleeSkillsTable:          reflect.TypeFor[CharacterMeleeSkill](),
	ClientStringsTable:                 reflect.TypeFor[ClientString](),
	ComponentChargesTable:              reflect.TypeFor[ComponentCharge](),
	CostTypesTable:                     reflect.TypeFor[CostType](),
	EffectivenessCostConstantsTable:    reflect.TypeFor[EffectivenessCostConstant](),
	EndgameMapsTable:                   reflect.TypeFor[EndgameMap](),
	FlasksTable:                        reflect.TypeFor[Flask](),
	FlavourTextTable:                   reflect.TypeFor[FlavourText](),
	GameConstantsTable:                 reflect.TypeFor[GameConstant](),
	GemEffectsTable:                    reflect.TypeFor[GemEffect](),
	GemTagsTable:                       reflect.TypeFor[GemTag](),
	GrantedEffectLabelsTable:           reflect.TypeFor[GrantedEffectLabel](),
	GrantedEffectQualityStatsTable:     reflect.TypeFor[GrantedEffectQualityStat](),
	GrantedEffectStatSetsPerLevelTable: reflect.TypeFor[GrantedEffectStatSetPerLevel](),
	GrantedEffectStatSetsTable:         reflect.TypeFor[GrantedEffectStatSet](),
	GrantedEffectsPerLevelTable:        reflect.TypeFor[GrantedEffectPerLevel](),
	GrantedEffectsTable:                reflect.TypeFor[GrantedEffect](),
	ItemClassCategoriesTable:           reflect.TypeFor[ItemClassCategory](),
	ItemClassesTable:                   reflect.TypeFor[ItemClass](),
	ItemExperiencePerLevelTable:        reflect.TypeFor[ItemExperiencePerLevel](),
	ItemExperienceTypesTable:           reflect.TypeFor[ItemExperienceType](),
	ItemInherentSkillsTable:            reflect.TypeFor[ItemInherentSkill](),
	ItemSpiritTable:                    reflect.TypeFor[ItemSpirit](),
	ModsTable:                          reflect.TypeFor[Mod](),
	MonsterCategoriesTable:             reflect.TypeFor[MonsterCategory](),
	MonsterPackEntriesTable:            reflect.TypeFor[MonsterPackEntry](),
	MonsterPacksTable:                  reflect.TypeFor[MonsterPack](),
	MonsterResistancesTable:            reflect.TypeFor[MonsterResistance](),
	MonsterTypesTable:                  reflect.TypeFor[MonsterType](),
	MonsterVarietiesTable:              reflect.TypeFor[MonsterVariety](),
	PlayerMinionIntrinsicStatsTable:    reflect.TypeFor[PlayerMinionIntrinsicStat](),
	QuestStaticRewardsTable:            reflect.TypeFor[QuestStaticReward](),
	ShieldTypesTable:                   reflect.TypeFor[ShieldType](),
	SkillGemSupportsTable:              reflect.TypeFor[SkillGemSupport](),
	SkillGemsTable:                     reflect.TypeFor[SkillGem](),
	SoulCoreLimitsTable:                reflect.TypeFor[SoulCoreLimit](),
	SoulCoreStatCategoriesTable:        reflect.TypeFor[SoulCoreStatCategory](),
	SoulCoreStatsTable:                 reflect.TypeFor[SoulCoreStat](),
	SoulCoreTypesTable:                 reflect.TypeFor[SoulCoreType](),
	SoulCoresTable:                     reflect.TypeFor[SoulCore](),
	StatsTable:                         reflect.TypeFor[Stat](),
	SupportGemFamilyTable:              reflect.TypeFor[SupportGemFamily](),
	SupportGemsTable:                   reflect.TypeFor[SupportGem](),
	TagsTable:                          reflect.TypeFor[Tag](),
	UtilityFlaskBuffsTable:             reflect.TypeFor[UtilityFlaskBuff](),
	WeaponTypesTable:                   reflect.TypeFor[WeaponType](),
	WieldableClassesTable:              reflect.TypeFor[WieldableClass](),
	WorldAreasTable:                    reflect.TypeFor[WorldArea](),
}
