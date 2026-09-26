package raw

// Every row and nested type is encoded as a strict tuple: fields by position, no keys. A struct change
// alters its loader.LayoutHash, so it is published under a new asset name instead of misreading old
// ones. Never vartuple, which decodes an inserted field into its neighbour's slot. TestTupleCoversAllStructs
// keeps this list complete.

//msgp:tuple ActiveSkill ActiveSkillType AlternatePassiveAddition AlternatePassiveSkill AlternateTreeVersion ArmourType AtlasNode AtlasNodeDefinition BaseItemType ComponentAttributeRequirement ComponentCharge Condition CostType CraftingBenchOption DefaultMonsterStat Essence Flask GrantedEffect GrantedEffectQualityStat GrantedEffectsPerLevel GrantedEffectStatSet GrantedEffectStatSetsPerLevel ItemClass ItemExperiencePerLevel LangTranslation Mod MonsterMapBossDifficulty MonsterMapDifficulty MonsterVariety PantheonPanelLayout PassiveSkill PassiveTreeExpansionJewel PassiveTreeExpansionSkill PassiveTreeExpansionSpecialSkill ShieldType SkillGem SkillTotem SkillTotemVariation Stat StatTranslation Tag TranslationFile WeaponType
