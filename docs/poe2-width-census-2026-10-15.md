# Row-width census (2026-10-15)

go-pob-data `354b991` + the width guard. **[M]** = measured, with the method given.

## The guard

`extractor.ParseDat` refuses a table when the schema row width differs from the data row width. The
data row width is `(boundary − 4) / rows`. The error names the table, both widths, the fixed bytes,
the row count and the schema file. pogo replaces the schema width with the data width
(`pogo/dat/parser.go:136–139`), so before this change a wrong layout decoded into wrong columns with
no error. The schema width is summed from the same `FieldType.Size` pogo uses, so it is the width pogo
would decode with.

`extractRawData` logs a refused table and skips it, as it does for any other parse error, so its
committed `json.gz` stays as it was. `treegen`'s `decodeTable` fails the run.

## PoE1 census: 84 of 84 present tables fit

Method [M]: a scratch tool ran `ParseDat` over every table in `raw.go:filesToExport` and every
`treegen` `loadTable` name (85 unique), on CDN patch `3.29.3.3` with `data/3.29/schema.min.json`.
That is the patch `ca04742` ("re-extract the 3.29 tables from patch 3.29.3.3") extracted from, and the
live patch on 2026-10-15. `SkillTotems` is absent from the bundle; it was already absent before.

Control [M]: the same tool, on the same data, with an older schema. The 3.28 schema misfits 9 of
those tables, and the 3.26 schema misfits 27. Examples: `BaseItemTypes` 310 vs 374,
`AlternatePassiveSkills` 168 vs 184. So the guard discriminates on real PoE1 data. The extractor's
live `TestParseDat` (31 tables) also passes with the guard in place.

**No PoE1 table go-pob-data emits is affected, so a re-extract would change nothing for go-pob.**

Not covered: the committed 3.18–3.28 data. The CDN serves only the live patch, so those versions
cannot be re-checked. Also not covered: a layout that has the correct total width but the wrong
column order or types. The guard cannot see that.

| table | status | rows | schema width | data width |
|---|---|---|---|---|
| ActiveSkills | fits | 1941 | 257 | 257 |
| ActiveSkillType | fits | 135 | 24 | 24 |
| AlternatePassiveAdditions | fits | 337 | 88 | 88 |
| AlternatePassiveSkills | fits | 381 | 184 | 184 |
| AlternateTreeArt | fits | 12 | 96 | 96 |
| AlternateTreeVersions | fits | 12 | 38 | 38 |
| ArmourTypes | fits | 481 | 60 | 60 |
| Ascendancy | fits | 21 | 141 | 141 |
| AtlasNode | fits | 182 | 261 | 261 |
| AtlasNodeDefinition | fits | 228 | 146 | 146 |
| BaseItemTypes | fits | 5461 | 374 | 374 |
| BlightCraftingItems | fits | 16 | 48 | 48 |
| BlightCraftingRecipes | fits | 576 | 56 | 56 |
| BlightCraftingResults | fits | 576 | 40 | 40 |
| BuffDefinitions | fits | 2814 | 314 | 314 |
| BuffTemplates | fits | 811 | 138 | 138 |
| Characters | fits | 7 | 612 | 612 |
| ClientStrings | fits | 8494 | 52 | 52 |
| ComponentAttributeRequirements | fits | 862 | 20 | 20 |
| ComponentCharges | fits | 51 | 24 | 24 |
| CostTypes | fits | 13 | 36 | 36 |
| CraftingBenchOptions | fits | 829 | 376 | 376 |
| CraftingItemClassCategories | fits | 22 | 56 | 56 |
| CurrencyExchange | fits | 1126 | 56 | 56 |
| DefaultMonsterStats | fits | 100 | 72 | 72 |
| Descendancy | fits | 17 | 124 | 124 |
| EffectivenessCostConstants | fits | 3 | 12 | 12 |
| Essences | fits | 106 | 989 | 989 |
| Flasks | fits | 51 | 92 | 92 |
| FlavourText | fits | 2962 | 20 | 20 |
| GameConstants | fits | 42 | 16 | 16 |
| GemEffects | fits | 1292 | 132 | 132 |
| GemTags | fits | 54 | 80 | 80 |
| GrantedEffectQualityStats | fits | 1014 | 80 | 80 |
| GrantedEffects | fits | 12829 | 235 | 235 |
| GrantedEffectsPerLevel | fits | 80600 | 120 | 120 |
| GrantedEffectStatSets | fits | 12763 | 68 | 68 |
| GrantedEffectStatSetsPerLevel | fits | 79927 | 184 | 184 |
| ItemClasses | fits | 103 | 153 | 153 |
| ItemExperiencePerLevel | fits | 361 | 24 | 24 |
| ItemVisualIdentity | fits | 15484 | 604 | 604 |
| ModFamily | fits | 7844 | 8 | 8 |
| Mods | fits | 40355 | 654 | 654 |
| ModType | fits | 14554 | 41 | 41 |
| MonsterMapBossDifficulty | fits | 90 | 104 | 104 |
| MonsterMapDifficulty | fits | 25 | 84 | 84 |
| MonsterResistances | fits | 50 | 56 | 56 |
| MonsterTypes | fits | 3600 | 48 | 48 |
| MonsterVarieties | fits | 10691 | 794 | 794 |
| PantheonPanelLayout | fits | 16 | 286 | 286 |
| PassiveJewelArt | fits | 16 | 48 | 48 |
| PassiveJewelSlots | fits | 60 | 77 | 77 |
| PassiveOverrideLimits | fits | 41 | 17 | 17 |
| PassiveSkillMasteryEffects | fits | 354 | 70 | 70 |
| PassiveSkillMasteryGroups | fits | 61 | 81 | 81 |
| PassiveSkillOverrides | fits | 343 | 156 | 156 |
| PassiveSkillOverrideTypes | fits | 24 | 26 | 26 |
| PassiveSkills | fits | 5706 | 354 | 354 |
| PassiveSkillTattoos | fits | 136 | 68 | 68 |
| PassiveSkillTattooTargetSets | fits | 10 | 40 | 40 |
| PassiveSkillTreeConnectionArt | fits | 1 | 64 | 64 |
| PassiveSkillTreeNodeFrameArt | fits | 12 | 48 | 48 |
| PassiveSkillTrees | fits | 7 | 75 | 75 |
| PassiveSkillTreeUIArtAscendancy | fits | 15 | 112 | 112 |
| PassiveTreeExpansionJewels | fits | 3 | 92 | 92 |
| PassiveTreeExpansionJewelSizes | fits | 3 | 8 | 8 |
| PassiveTreeExpansionSkills | fits | 55 | 64 | 64 |
| PassiveTreeExpansionSpecialSkills | fits | 309 | 32 | 32 |
| ReminderText | fits | 468 | 24 | 24 |
| ShieldTypes | fits | 98 | 20 | 20 |
| SkillGems | fits | 847 | 135 | 135 |
| SkillTotems | absent from the 3.29.3.3 bundle | | | |
| SkillTotemVariations | fits | 285 | 24 | 24 |
| Stats | fits | 23346 | 105 | 105 |
| Tags | fits | 1389 | 28 | 28 |
| Tinctures | fits | 10 | 24 | 24 |
| UIArtAscendancy | fits | 15 | 116 | 116 |
| UniqueStashLayout | fits | 1556 | 83 | 83 |
| UniqueStashTypes | fits | 23 | 53 | 53 |
| VillageBalancePerLevelShared | fits | 100 | 28 | 28 |
| VirtualStatContextFlags | fits | 40 | 12 | 12 |
| WeaponPassiveSkills | fits | 2632 | 100 | 100 |
| WeaponPassiveSkillTypes | fits | 3 | 12 | 12 |
| WeaponTypes | fits | 366 | 40 | 40 |
| Words | fits | 4117 | 64 | 64 |

## PoE2: the four "misfits" were an unread `interval` flag

dat-schema marks some `i32` columns `"interval": true`. Such a column is a pair, a min and a max, so
it is 8 bytes. `extractor.Column` had no `Interval` field, so every interval was read as 4 bytes. In
the 2026-10-14 schema [M], all 21 interval columns belong to the `validFor: 2` (PoE2) entries of
exactly four tables: `Mods` (8), `AlternatePassiveSkills` (7), `AlternatePassiveAdditions` (3) and
`AlternateTreeVersions` (3). The 4 bytes missing per column account for each research-lane misfit
exactly: 661 + 32 = 693, 140 + 28 = 168, 76 + 12 = 88, and 26 + 12 = 38. **dat-schema did not lag
the patch.** That refutes risk 2 ("schema lag") of `poe2-data-2026-10-15.md` for these tables. No
PoE1 or shared table has an interval column, so this changes nothing for PoE1 (the census above was
re-run after the fix and is identical).

The fix: `ToJSONFormat` emits an interval column as two fields of its type, `<Name>Min` and
`<Name>Max`. pogo has no interval type.

Proof [M], on PoE2 patch `0.5.5.4.2` (`patch-poe2.poecdn.com`, `data/balance/`, new hash), with the
latest dat-schema reduced to `validFor 2` where both entries exist:

- With the fix, all four tables fit: `Mods` 693/693 (16 784 rows), `AlternatePassiveSkills` 168/168,
  `AlternatePassiveAdditions` 88/88 and `AlternateTreeVersions` 38/38. `Stats` and `CostTypes` also
  fit as controls.
- With the interval handling mutated off, the guard refuses all four, at the research lane's exact
  widths (661, 140, 76, 26). `Stats` still fits.
- Spot-check against PoB-PoE2 `src/Data/ModItem.lua`: `Strength1` gives level 1 and
  `additional_strength` 5–8, which matches "+(5-8) to Strength". `LocalAddedPhysicalDamage6` gives
  level 54 and min/max added physical damage (13–20) and (23–35), which matches. `FishingRarity`
  gives `fish_rarity_+%` 25–40, which matches.
  PoB-PoE2 does not export the timeless tables, so they are checked by width and by a plausible first
  row (`vaal_keystone_1` "Divine Flesh").

PoE2 is still not extractable end to end. Game plumbing is not done: `LoadSchema` keeps `validFor 1`,
`IsNewHash` and `GetFileExtension` misread "0.x", and `Data/Balance/` and the CDN host are not
handled.

## Per-table schema override: not built

There is nothing to fill one with. All present PoE1 tables fit dat-schema. The four PoE2 misfits were
the interval bug and now fit dat-schema unmodified. An override mechanism with no entries would be
dead code. Build it when the guard first refuses a table that dat-schema has not yet caught up with.
The guard's error names the table and both widths, which is the input such an override needs.
PoB-PoE2's `spec.lua` misfits 5 other PoE2 tables (research lane), so it is not a drop-in override
source.
