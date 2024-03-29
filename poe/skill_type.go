package poe

type SkillType string

const (
	SkillTypeAttack                         = SkillType("Attack")
	SkillTypeSpell                          = SkillType("Spell")
	SkillTypeProjectile                     = SkillType("Projectile")    // Specifically skills which fire projectiles
	SkillTypeDualWieldOnly                  = SkillType("DualWieldOnly") // Attack requires dual wielding only used on Dual Strike
	SkillTypeBuff                           = SkillType("Buff")
	SkillTypeRemoved6                       = SkillType("Removed6")     // Now removed was CanDualWield: Attack can be used while dual wielding
	SkillTypeMainHandOnly                   = SkillType("MainHandOnly") // Attack only uses the main hand; removed in 3.5 but still needed for 2.6
	SkillTypeRemoved8                       = SkillType("Removed8")     // Now removed was only used on Cleave
	SkillTypeMinion                         = SkillType("Minion")
	SkillTypeDamage                         = SkillType("Damage") // Skill hits (not used on attacks because all of them hit)
	SkillTypeArea                           = SkillType("Area")
	SkillTypeDuration                       = SkillType("Duration")
	SkillTypeRequiresShield                 = SkillType("RequiresShield")
	SkillTypeProjectileSpeed                = SkillType("ProjectileSpeed")
	SkillTypeHasReservation                 = SkillType("HasReservation")
	SkillTypeReservationBecomesCost         = SkillType("ReservationBecomesCost")
	SkillTypeTrappable                      = SkillType("Trappable")       // Skill can be turned into a trap
	SkillTypeTotemable                      = SkillType("Totemable")       // Skill can be turned into a totem
	SkillTypeMineable                       = SkillType("Mineable")        // Skill can be turned into a mine
	SkillTypeElementalStatus                = SkillType("ElementalStatus") // Causes elemental status effects but doesn't hit (used on Herald of Ash to allow Elemental Proliferation to apply)
	SkillTypeMinionsCanExplode              = SkillType("MinionsCanExplode")
	SkillTypeRemoved22                      = SkillType("Removed22") // Now removed was AttackCanTotem
	SkillTypeChains                         = SkillType("Chains")
	SkillTypeMelee                          = SkillType("Melee")
	SkillTypeMeleeSingleTarget              = SkillType("MeleeSingleTarget")
	SkillTypeMulticastable                  = SkillType("Multicastable") // Spell can repeat via Spell Echo
	SkillTypeTotemCastsAlone                = SkillType("TotemCastsAlone")
	SkillTypeMultistrikeable                = SkillType("Multistrikeable") // Attack can repeat via Multistrike
	SkillTypeCausesBurning                  = SkillType("CausesBurning")   // Deals burning damage
	SkillTypeSummonsTotem                   = SkillType("SummonsTotem")
	SkillTypeTotemCastsWhenNotDetached      = SkillType("TotemCastsWhenNotDetached")
	SkillTypeFire                           = SkillType("Fire")
	SkillTypeCold                           = SkillType("Cold")
	SkillTypeLightning                      = SkillType("Lightning")
	SkillTypeTriggerable                    = SkillType("Triggerable")
	SkillTypeTrapped                        = SkillType("Trapped")
	SkillTypeMovement                       = SkillType("Movement")
	SkillTypeRemoved39                      = SkillType("Removed39") // Now removed was Cast
	SkillTypeDamageOverTime                 = SkillType("DamageOverTime")
	SkillTypeRemoteMined                    = SkillType("RemoteMined")
	SkillTypeTriggered                      = SkillType("Triggered")
	SkillTypeVaal                           = SkillType("Vaal")
	SkillTypeAura                           = SkillType("Aura")
	SkillTypeRemoved45                      = SkillType("Removed45")               // Now removed was LightningSpell
	SkillTypeCanTargetUnusableCorpse        = SkillType("CanTargetUnusableCorpse") // Doesn't appear to be used at all
	SkillTypeRemoved47                      = SkillType("Removed47")               // Now removed was TriggeredAttack
	SkillTypeRangedAttack                   = SkillType("RangedAttack")
	SkillTypeRemoved49                      = SkillType("Removed49") // Now removed was MinionSpell
	SkillTypeChaos                          = SkillType("Chaos")
	SkillTypeFixedSpeedProjectile           = SkillType("FixedSpeedProjectile") // Not used by any skill
	SkillTypeRemoved52                      = SkillType("Removed52")
	SkillTypeThresholdJewelArea             = SkillType("ThresholdJewelArea") // Allows Burning Arrow and Vigilant Strike to be supported by Inc AoE and Conc Effect
	SkillTypeThresholdJewelProjectile       = SkillType("ThresholdJewelProjectile")
	SkillTypeThresholdJewelDuration         = SkillType("ThresholdJewelDuration") // Allows Burning Arrow to be supported by Inc/Less Duration and Rapid Decay
	SkillTypeThresholdJewelRangedAttack     = SkillType("ThresholdJewelRangedAttack")
	SkillTypeRemoved57                      = SkillType("Removed57")
	SkillTypeChannel                        = SkillType("Channel")
	SkillTypeDegenOnlySpellDamage           = SkillType("DegenOnlySpellDamage") // Allows Contagion Blight and Scorching Ray to be supported by Controlled Destruction
	SkillTypeRemoved60                      = SkillType("Removed60")            // Now removed was ColdSpell
	SkillTypeInbuiltTrigger                 = SkillType("InbuiltTrigger")       // Skill granted by item that is automatically triggered prevents trigger gems and trap/mine/totem from applying
	SkillTypeGolem                          = SkillType("Golem")
	SkillTypeHerald                         = SkillType("Herald")
	SkillTypeAuraAffectsEnemies             = SkillType("AuraAffectsEnemies") // Used by Death Aura added by Blasphemy
	SkillTypeNoRuthless                     = SkillType("NoRuthless")
	SkillTypeThresholdJewelSpellDamage      = SkillType("ThresholdJewelSpellDamage")
	SkillTypeCascadable                     = SkillType("Cascadable")                     // Spell can cascade via Spell Cascade
	SkillTypeProjectilesFromUser            = SkillType("ProjectilesFromUser")            // Skill can be supported by Volley
	SkillTypeMirageArcherCanUse             = SkillType("MirageArcherCanUse")             // Skill can be supported by Mirage Archer
	SkillTypeProjectileSpiral               = SkillType("ProjectileSpiral")               // Excludes Volley from Vaal Fireball and Vaal Spark
	SkillTypeSingleMainProjectile           = SkillType("SingleMainProjectile")           // Excludes Volley from Spectral Shield Throw
	SkillTypeMinionsPersistWhenSkillRemoved = SkillType("MinionsPersistWhenSkillRemoved") // Excludes Summon Phantasm on Kill from Manifest Dancing Dervish
	SkillTypeProjectileNumber               = SkillType("ProjectileNumber")               // Allows LMP/GMP on Rain of Arrows and Toxic Rain
	SkillTypeWarcry                         = SkillType("Warcry")                         // Warcry
	SkillTypeInstant                        = SkillType("Instant")                        // Instant cast skill
	SkillTypeBrand                          = SkillType("Brand")
	SkillTypeDestroysCorpse                 = SkillType("DestroysCorpse") // Consumes corpses on use
	SkillTypeNonHitChill                    = SkillType("NonHitChill")
	SkillTypeChillingArea                   = SkillType("ChillingArea")
	SkillTypeAppliesCurse                   = SkillType("AppliesCurse")
	SkillTypeCanRapidFire                   = SkillType("CanRapidFire")
	SkillTypeAuraDuration                   = SkillType("AuraDuration")
	SkillTypeAreaSpell                      = SkillType("AreaSpell")
	SkillTypeOR                             = SkillType("OR")
	SkillTypeAND                            = SkillType("AND")
	SkillTypeNOT                            = SkillType("NOT")
	SkillTypePhysical                       = SkillType("Physical")
	SkillTypeAppliesMaim                    = SkillType("AppliesMaim")
	SkillTypeCreatesMinion                  = SkillType("CreatesMinion")
	SkillTypeGuard                          = SkillType("Guard")
	SkillTypeTravel                         = SkillType("Travel")
	SkillTypeBlink                          = SkillType("Blink")
	SkillTypeCanHaveBlessing                = SkillType("CanHaveBlessing")
	SkillTypeProjectilesNotFromUser         = SkillType("ProjectilesNotFromUser")
	SkillTypeAttackInPlaceIsDefault         = SkillType("AttackInPlaceIsDefault")
	SkillTypeNova                           = SkillType("Nova")
	SkillTypeInstantNoRepeatWhenHeld        = SkillType("InstantNoRepeatWhenHeld")
	SkillTypeInstantShiftAttackForLeftMouse = SkillType("InstantShiftAttackForLeftMouse")
	SkillTypeAuraNotOnCaster                = SkillType("AuraNotOnCaster")
	SkillTypeBanner                         = SkillType("Banner")
	SkillTypeRain                           = SkillType("Rain")
	SkillTypeCooldown                       = SkillType("Cooldown")
	SkillTypeThresholdJewelChaining         = SkillType("ThresholdJewelChaining")
	SkillTypeSlam                           = SkillType("Slam")
	SkillTypeStance                         = SkillType("Stance")
	SkillTypeNonRepeatable                  = SkillType("NonRepeatable") // Blood and Sand + Flesh and Stone
	SkillTypeOtherThingUsesSkill            = SkillType("OtherThingUsesSkill")
	SkillTypeSteel                          = SkillType("Steel")
	SkillTypeHex                            = SkillType("Hex")
	SkillTypeMark                           = SkillType("Mark")
	SkillTypeAegis                          = SkillType("Aegis")
	SkillTypeOrb                            = SkillType("Orb")
	SkillTypeKillNoDamageModifiers          = SkillType("KillNoDamageModifiers")
	SkillTypeRandomElement                  = SkillType("RandomElement") // means elements cannot repeat
	SkillTypeLateConsumeCooldown            = SkillType("LateConsumeCooldown")
	SkillTypeArcane                         = SkillType("Arcane") // means it is reliant on amount of mana spent
	SkillTypeFixedCastTime                  = SkillType("FixedCastTime")
	SkillTypeRequiresOffHandNotWeapon       = SkillType("RequiresOffHandNotWeapon")
	SkillTypeLink                           = SkillType("Link")
	SkillTypeBlessing                       = SkillType("Blessing")
	SkillTypeZeroReservation                = SkillType("ZeroReservation")
)
