# Core Mechanics Updates for Lunar VI

## Overview
This document outlines critical core modifications introduced to support the Lunar VI custom mechanics (Lunar Reactions, Hexerei traits, etc.).

## 1. Attack Modification System (`mods.go`)

### Critical Update: Lunar Reaction Tags
**Date**: 2026-01-25
**Affected File**: `pkg/core/player/character/mods.go`

The `ApplyAttackMods` function in the core system has been modified to explicitly support **Lunar Reaction Tags**.

#### Problem
By default, gcsim's core logic (`ApplyAttackMods`) filters out any attack with a tag value greater than or equal to `AttackTagNoneStat` (the threshold for reaction types). This prevents standard reactions from receiving stat buffs (ATK%, DMG%, etc.) through AttackMods, which is the intended behavior for vanilla Genshin Impact reactions (Swirl, Bloom, etc. rely on EM/Level, not ATK/CRIT).

However, **Lunar Reactions** (`LCDamage` (Lunar-Charged), `LBDamage` (Lunar-Bloom), `LCrsDamage` (Lunar-Crystallize)) are hybrid reactions that:
1. Are classified as reaction tags (so they don't break other logic).
2. **DO** scale with character stats (ATK, CRIT, DMG Bonus).

Previously, this filtering logic caused Lunar Reactions (and abilities inheriting these tags, like Columbina's `Gravity Interference`) to reject all AttackMods (such as weapon passives like *Nocturne's Curtain Call*).

#### Solution
The filter condition in `ApplyAttackMods` has been updated to **whitelist** Lunar Reaction tags:

```go
// Old Logic
if a.Info.AttackTag >= attacks.AttackTagNoneStat {
    return
}

// New Logic
if a.Info.AttackTag >= attacks.AttackTagNoneStat && 
   a.Info.AttackTag != attacks.AttackTagLCDamage &&
   a.Info.AttackTag != attacks.AttackTagLBDamage &&
   a.Info.AttackTag != attacks.AttackTagLCrsDamage {
    return
}
```

### Stat Restriction for Lunar Reactions
**Date**: 2026-01-25 (Updated)

While Lunar Reactions are whitelisted to enter `ApplyAttackMods`, they are strictly restricted to receiving **only CRIT Rate (CR) and CRIT DMG (CD)** buffs from these mods.

```go
// Inside ApplyAttackMods loop
if isLunar && attributes.Stat(k) != attributes.CR && attributes.Stat(k) != attributes.CD {
    continue
}
```

#### Implications
- Lunar Reactions will profit from dynamic CR/CD buffs (e.g. Blizzard Strayer 4pc, Rosaria A4).
- They will **NOT** receive dynamic ATK% or DMG% buffs from AttackMods (these must be snapshot or applied via other means if intended).
- This ensures that weapons and artifacts providing `AddAttackMod` buffs will correctly apply CRIT stats to Lunar mechanics without unintended ATK bloat.
