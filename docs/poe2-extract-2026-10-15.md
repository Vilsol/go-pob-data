# PoE2 extraction (2026-10-15)

go-pob-data now extracts PoE2 (`-game poe2`) beside PoE1 (`-game poe1`, the default). **[M]** = measured, with
the method given.

## Usage

```bash
go run . -game poe2 -latest                       # prints the live patch, e.g. 0.5.5.4.2
go run . -game poe2 cdn:0.5.5.4.2 0.5             # extract into ./data/0.5 (cgo; needs data/0.5/schema.min.json)
./extract.sh cdn:0.5.5.4.2 0.5 poe2               # downloads the schema first
mise run extract poe2                             # the live patch, through extract.sh
```

The binary writes under `data/` of its working directory, so a scratch run is the binary run from a scratch
directory that holds `data/<version>/schema.min.json`.

## What changed

| area | PoE1 | PoE2 |
|---|---|---|
| bundle hash, dat extension | minor ≥ 22 new hash, ≥ 25 `.datc64` | always new hash and `.datc64` (`extractor.IsNewHash` / `GetFileExtension` take the game) |
| table directory | `Data/` | `Data/Balance/` (`Game.TableDir`) |
| CDN | `patch.poecdn.com` | `patch-poe2.poecdn.com` |
| live patch | `patch.pathofexile.com:12995` | `patch.pathofexile2.com:13060` (`extractor.LatestPatch`, `-latest`) |
| dat-schema entry | `validFor` 1 or 3 | `validFor` 2 or 3 (the other game's entries are dropped) |
| tables | `filesToExport` (71, json + tuple) | `poe2TablesToExport` (96, json.gz + json.br only) |
| stat descriptions | `Metadata/StatDescriptions/*.txt` (`st.Files`) | `Data/StatDescriptions/*.csd`: `st.PoE2Files` (27) plus every file under `specific_skill_stat_descriptions/` (562), found in the game |
| text files | `textFilesToExport` | the same less `skillpopup_stat_filters.txt`; monster list `monster_varieties_poe2.txt` from PoB-PoE2 |
| images | 6 fixed `BaseClassIllustrations` | the `PassiveTreeImage` of every `Characters` row (12) |
| tree | `treegen` | skipped; `-tree-only` refuses |

Parser forms added for PoE2 [M]: `include "Data/StatDescriptions/<f>.csd"`, and a repeated language block. PoE2's
`atlas_variant_stat_descriptions.csd:266` repeats `lang "Portuguese"` with the same line. A repeat is accepted only
when its lines equal the first block's; a different repeat, or a repeated English block, still fails.

PoE2 skips two things PoE1 refuses, each with a warning. The first is a monster variety that PoB-PoE2's templates
name but the game does not have: `BloodClan/BloodClanDagger` [M]. The second is a template that a variety names
but the game does not ship: `BloodClan/BloodClan.ot`, the base of `BloodClanVillager`. PoB-PoE2 skips such a
template too (`Export/Scripts/minions.lua:55-57`).

## End-to-end run [M]

PoE2 patch `0.5.5.4.2` with dat-schema `latest` (version 7, createdAt 1791971088). The run used a fresh bundle cache
and a scratch working directory, at load average 4.1. Exit 0, no `ERROR`, and the 2 warnings above.

- Wall time: **38 s**, including the download.
- Downloaded: **410 211 455 B** (84 files: the bundle index and the bundles read).
- Tables: **96** `json.gz` + 96 `json.br`, **231 048 rows**. Every table passes the width guard, so no override
  was needed. 6 tables are empty: `AtlasNode`, `AtlasNodeDefinition`, `PassiveSkillOverrides`, and
  `PassiveTreeExpansion{Jewels,Skills,SpecialSkills}`. Largest: `GrantedEffectStatSetsPerLevel` 41 557,
  `GrantedEffectsPerLevel` 34 300, `Stats` 27 281, `ItemVisualIdentity` 17 227, `Mods` 16 784.
- Translations: **589 files × 11 languages** (cn de en es fr jp kr po ru th tw), 27 top-level and 562 specific.
  Every one of the 589 `.csd` files parses. Control: before the repeat fix, the same census failed exactly 1.
- Files: **685** (595 `.ot`, 89 `.it`, `art/uiimages1.txt`).
- Images: **12** PNG.
- Output: 44.7 MB (raw 23.4 MB, stat_translations 18.3 MB, files 0.4 MB).

Rows per table: `ActiveSkills` 1147, `ActiveSkillType` 290, `AlternatePassiveAdditions` 95,
`AlternatePassiveSkills` 231, `AlternateTreeArt` 8, `AlternateTreeVersions` 8, `ArmourTypes` 1280, `Ascendancy` 37,
`AscendancyPassiveSkillOverrides` 13, `AttributeRequirements` 1735, `BaseItemTypes` 5496, `BlightCraftingRecipes`
1017, `BlightCraftingResults` 1017, `BuffTemplates` 665, `BuildPlannerInventories` 17, `CharacterMeleeSkills` 38,
`Characters` 12, `ClassPassiveSkillOverrides` 61, `ClientStrings` 9640, `ComponentCharges` 31, `CostTypes` 18,
`CraftingBenchOptions` 85, `CraftingItemClassCategories` 29, `CurrencyExchange` 687, `DefaultMonsterStats` 100,
`EffectivenessCostConstants` 3, `EndgameMaps` 173, `EssenceMods` 178, `Essences` 95, `Expedition2VerisiumCrafts` 718,
`Flasks` 31, `FlatPhysicalDamageValues` 40, `FlavourText` 3591, `GameConstants` 121, `GemEffects` 1194, `GemTags` 67,
`GoldRespecPrices` 100, `GrantedEffectQualityStats` 429, `GrantedEffects` 8349, `GrantedEffectsPerLevel` 34300,
`GrantedEffectStatSets` 8283, `GrantedEffectStatSetsPerLevel` 41557, `ItemClasses` 118, `ItemExperiencePerLevel` 375,
`ItemInherentSkills` 253, `ItemSpirit` 16, `ItemVisualIdentity` 17227, `LiquidEmotionOutcomes` 26,
`MinionGemLevelScaling` 40, `ModFamily` 6126, `ModGrantedSkills` 65, `Mods` 16784, `ModType` 12574,
`MonsterCategories` 6, `MonsterMapBossDifficulty` 90, `MonsterMapDifficulty` 25, `MonsterPackEntries` 2085,
`MonsterPacks` 1116, `MonsterResistances` 22, `MonsterTypes` 1571, `MonsterVarieties` 2794, `PantheonPanelLayout` 16,
`PassiveJewelArt` 9, `PassiveJewelSlots` 31, `PassiveJewelUniqueArt` 11, `PassiveOverrideLimits` 7,
`PassiveSkillOverrideTypes` 1, `PassiveSkills` 9731, `PassiveSkillTreeConnectionArt` 10,
`PassiveSkillTreeNodeFrameArt` 45, `PassiveSkillTrees` 5, `PassiveTreeExpansionJewelSizes` 3,
`PlayerMinionIntrinsicStats` 4, `ShieldTypes` 203, `SkillGems` 1191, `SkillGemSupports` 410, `SkillTotemVariations`
24, `SoulCores` 313, `SoulCoreStats` 526, `SpectreOverrides` 206, `Stats` 27281, `SupportGems` 637, `Tags` 1339,
`UniqueOrigins` 126, `UniqueStashLayout` 449, `UniqueStashTypes` 34, `VirtualStatContextFlags` 42, `WeaponTypes` 407,
`Words` 3246, `WorldAreas` 442; the 6 empty tables are listed above.

The PoE2 list is `filesToExport` less the 8 tables PoE2 does not ship (63), plus 33 more. 32 of those are what
PoB-PoE2's `src/Export` reads and go-pob-data did not export (the research note counted 28 because it left out
`Ascendancy`, `BlightCraftingRecipes`, `BlightCraftingResults` and `PassiveSkillTrees`, which PoE1 reads only in
treegen). The 33rd is `Characters`, the source of the class illustrations.

## PoE1 is unchanged [M]

The changed code re-extracted `Mods`, `Stats` and `CostTypes` from CDN `3.29.3.3` with the committed
`data/3.29/schema.min.json` into scratch. Their `json.gz`, `json.br` and current tuple asset are byte-identical
to the committed `data/3.29` files (9 of 9). The same run re-extracted `stat_descriptions` and
`skill_stat_descriptions` in all languages, and every text file and template. That is 379 files, all
byte-identical. Control: with the schema pick forced to PoE2's entries, the width guard refuses all 3 tables and
the proof fails. The live `extractor`, `treegen` and root tests pass, `TestDiffAgainstExport` among them.

## Left for later

- **#6 raw structs and tuple assets.** PoE2 tables are json only. The PoE1 structs must never encode them,
  because 39 tables have a different PoE2 layout and `json.Unmarshal` would drop columns silently.
- **#7 PoE2 treegen.** Not run for PoE2.
- **Manifest and meta.** `manifestJSON` lists only top-level translations (`en/*.json.gz`), so the 562 files
  under `specific_skill_stat_descriptions/` would be published but not listed. `generateMeta` puts `0.5` in the
  flat `versions` list beside `3.x`. Both have to be settled before any PoE2 version is committed.
- **Includes.** Each translation keeps its `include` lines as the game writes them
  (`Data/StatDescriptions/<f>.csd`). A consumer that resolves includes has to map that path for PoE2.
- **Class list.** All 12 `Characters` rows are class starts in PoE2's `PassiveSkills`, so all 12 illustrations
  are exported. That includes `Marauder`, `Duelist`, `Shadow` and `Templar`, which have PoE2-specific art.
  PoB-PoE2's tree filters only on `[DNT` names (`passivetree.lua:15`), and no row has one.
- **Schema drift.** The run used the latest dat-schema. The guard is the only check that a later PoE2 patch
  still fits it.
