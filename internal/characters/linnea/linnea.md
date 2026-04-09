# linnea (ID:90660 / JP:リンネア)

Region: 12
Geo / Bow
Rarity: Orange

skillcon: 3
burstcon: 5

**注意 必ず参照すること**
- 天賦倍率の配列はすべて、必ず15エントリ作ってください。数値がない場合、最後の数値を15エントリになるまで繰り返すこと。
- 同名のStatusやStatMod,AttackModは同じものとして処理され、上書きされるため注意すること。
- 必要に応じて、**許可されたwebサイトでのみ**情報を収集してかまわない。サイトはcopilot-instructions.mdを参照のこと。

---

## Ascension Phase (data_gen.textproto用)

| Level | FIGHT_PROP_BASE_HP | FIGHT_PROP_BASE_ATTACK | FIGHT_PROP_BASE_DEFENSE | FIGHT_PROP_CRITICAL |
|-------|---------|---------|---------|----------|
| **0** |
| 1/20 | 770 | 11 | 71 | 0.0% |
| 20/20 | 1998 | 29 | 183 | 0.0% |
| **1** |
| 20/40 | 2659 | 39 | 244 | 0.0% |
| 40/40 | 3978 | 58 | 365 | 0.0% |
| **2** |
| 40/50 | 4447 | 64 | 408 | 4.8% |
| 50/50 | 5117 | 74 | 469 | 4.8% |
| **3** |
| 50/60 | 5742 | 83 | 526 | 9.6% |
| 60/60 | 6419 | 93 | 588 | 9.6% |
| **4** |
| 60/70 | 6888 | 100 | 631 | 9.6% |
| 70/70 | 7570 | 110 | 694 | 9.6% |
| **5** |
| 70/80 | 8040 | 117 | 737 | 14.4% |
| 80/80 | 8730 | 127 | 800 | 14.4% |
| **6** |
| 80/90 | 9199 | 133 | 843 | 19.2% |
| 90/90 | 9895 | 143 | 907 | 19.2% |

---

## Normal Attack - Capture Protocol

Normal Attack
Performs up to 3 consecutive shots with a bow.

Charged Attack
Performs a more precise Aimed Shot with increased DMG.
While aiming, stone crystals will accumulate on the arrowhead. A fully charged crystalline arrow will deal Geo DMG.

Plunging Attack
Fires off a shower of arrows in mid-air before falling and striking the ground, dealing AoE DMG upon impact.
 
> **実装ノート:** ほかのBowキャラと共通。

### Talent Scaling
Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11
1-Hit DMG 	59.0%	63.8%	68.6%	75.5%	80.3%	85.8%	93.3%	100.8%	108.4%	116.6%	124.9%
2-Hit DMG 	51.2%	55.3%	59.5%	65.4%	69.6%	74.3%	80.9%	87.4%	94.0%	101.1%	108.3%
3-Hit DMG 	81.6%	88.3%	94.9%	104.4%	111.1%	118.7%	129.1%	139.5%	150.0%	161.4%	172.8%
Aimed Shot 	43.9%	47.4%	51.0%	56.1%	59.7%	63.7%	69.4%	75.0%	80.6%	86.7%	92.8%
Fully-Charged Aimed Shot 	124%	133%	143%	155%	164%	174%	186%	198%	211%	223%	236%
Plunge DMG 	56.8%	61.5%	66.1%	72.7%	77.3%	82.6%	89.9%	97.1%	104.4%	112.3%	120.3%
Low/High Plunge DMG 	114%/142%	123%/153%	132%/165%	145%/182%	155%/193%	165%/206%	180%/224%	194%/243%	209%/261%	225%/281%	240%/300%

> **実装ノート:** Physical攻撃。ICDはほかのBowキャラと共通。

---

## Elemental Skill - Windbound Execution

Tap
Lumi strikes in a "Super Power Form", attacking nearby enemies continuously, dealing AoE Geo DMG. If there are Moondrifts nearby, Lumi will also deal AoE Geo DMG to nearby enemies, which will be considered Lunar-Crystallize Reaction DMG.
> **実装ノート:** ”considered Lunar-Crystallize Reaction DMG” はzibaiなどの実装を参考にすること。

Continuous Tapping
Lumi will strike in her "Ultimate Power Form", dealing an especially powerful instance of AoE Geo DMG that is considered Lunar-Crystallize Reaction DMG, switching to "Standard Power Form".
Additionally, tapping the Elemental Skill increases Linnea's interruption resistance.
> **実装ノート:** gcsl上では　skill["mash=1"]　で使用する。

Super Power Form
Lumi will use Pound-Pound Pummeler to attack nearby opponents continuously. Every attack will deal 2 instances of AoE Geo DMG.
If there are Moondrifts nearby, Lumi will alternate between 2 Pound-Pound Pummeler and 1 Heavy Overdrive Hammer to continuously attack nearby opponents. Heavy Overdrive Hammer will deal AoE Geo DMG which will be considered Lunar-Crystallize Reaction DMG.

Ultimate Power Form
Lumi will use the ultimate Million Ton Crush, dealing an especially powerful instance of AoE Geo DMG that is considered Lunar-Crystallize Reaction DMG, switching to Standard Power Form.

Standard Power Form
Lumi will use Pound-Pound Pummeler at longer intervals to attack nearby opponents continuously. Every attack will deal 2 instances of AoE Geo DMG.

When this skill hits at least one enemy, it generates 3 Elemental Particles.

### Talent Scaling
Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11	Lv.12	Lv.13
Lumi Pound-Pound Pummeler DMG 	96.0% DEF ×2	103.2% DEF ×2	110.4% DEF ×2	120.0% DEF ×2	127.2% DEF ×2	134.4% DEF ×2	144.0% DEF ×2	153.6% DEF ×2	163.2% DEF ×2	172.8% DEF ×2	182.4% DEF ×2	192.0% DEF ×2	204.0% DEF ×2
Lumi Heavy Overdrive Hammer DMG 	100.0% DEF	107.5% DEF	115.0% DEF	125.0% DEF	132.5% DEF	140.0% DEF	150.0% DEF	160.0% DEF	170.0% DEF	180.0% DEF	190.0% DEF	200.0% DEF	212.5% DEF
Lumi Million Ton Crush DMG 	400% DEF	430% DEF	460% DEF	500% DEF	530% DEF	560% DEF	600% DEF	640% DEF	680% DEF	720% DEF	760% DEF	800% DEF	850% DEF
Lumi Duration 	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s	25.0s
CD 	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s	18.0s

Lumi Pound-Pound Pummeler DMG: Geo / Durability 25 / Tag: ICDTagElementalArt Skill / Group: ICDGroupDefault
Lumi Heavy Overdrive Hammer DMG: Geo(Lunar-Crystallize) / Durability 0 / No ICD
Lumi Million Ton Crush DMG: Geo(Lunar-Crystallize) / Durability 0 / No ICD

---

## Elemental Burst - Memo: Survival Guide in Extreme Conditions

Linnea summons Lumi to strike in Super Power Form, healing nearby party members. For a short duration, she will continuously heal nearby active party members based on Linnea's DEF.
If Lumi is already on the field when Linnea unleashes Elemental Burst, Lumi's active duration will be reset instead and her Strike Form will not change.

### Talent Scaling
Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11	Lv.12	Lv.13
Initial Healing Amount 	160.0% DEF+770	172.0% DEF+847	184.0% DEF+930	200.0% DEF+1020	212.0% DEF+1117	224.0% DEF+1219	240.0% DEF+1328	256.0% DEF+1444	272.0% DEF+1566	288.0% DEF+1694	304.0% DEF+1829	320.0% DEF+1971	340.0% DEF+2118
Continuous Healing 	32.0% DEF+154	34.4% DEF+169	36.8% DEF+186	40.0% DEF+204	42.4% DEF+223	44.8% DEF+243	48.0% DEF+265	51.2% DEF+288	54.4% DEF+313	57.6% DEF+338	60.8% DEF+365	64.0% DEF+394	68.0% DEF+423
Healing Duration 	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s
CD 	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s
Energy Cost 	60	60	60	60	60	60	60	60	60	60	60	60	60


---

## Ascension Passive A1 - Field Observation Notes

When Lumi is present on the field, the Geo RES of opponents near Lumi will decrease by 15%.

Moonsign: Ascendant Gleam: Linnea's Elemental Skill Countermeasure: Lumi's Battle Cry! and Elemental Burst Memo: Survival Guide in Extreme Conditions are enhanced. After summoning Lumi, the Geo RES of opponents near Lumi will be further decreased by 15%.

---

## Ascension Passive A4 - Universal Naturalist Archive

Linnea will increase the Elemental Mastery of certain characters in your party according to your current active character. Increase in Elemental Mastery is based on 5% of Linnea's DEF. If your current active character:
· Is a Moonsign character: Increase this character's Elemental Mastery.
· Is not a Moonsign character: Increase Linnea's own Elemental Mastery.

---

## Ascension Passive　A0: Moonsign Benediction: Habitat Survey

When a party member triggers a Hydro Crystallize reaction, it will be converted into the Lunar-Crystallize reaction, with every 100 DEF that Linnea has increasing Lunar-Crystallize's Base DMG by 0.7%, up to a maximum of 14%.

Additionally, when Linnea is in the party, the party's Moonsign will increase by 1 level.

> **実装ノート:** ZibaiのA0と同じ効果。

---

## Constellation C1 - Provisional Classification
When unleashing Elemental Skill Countermeasure: Lumi's Battle Cry!, or when triggering Moondrift Harmony, Linnea gains 6 stacks of the Field Catalog effect for 10s. Max 18 stacks. When nearby party members deal Lunar-Crystallize Reaction DMG, consume 1 stack of Field Catalog to increase the DMG dealt. The increase in DMG is equal to 75% of Linnea's DEF.
Additionally, when Lumi uses Million Ton Crush in her Ultimate Power Form, Linnea can consume up to 5 stacks of Field Catalog. Each stack will increase the DMG dealt by 150% of Linnea's DEF.

---

## Constellation C2 - Tidings of Joy and Sorrow

Within 8s after triggering Moondrift Harmony, all Hydro and Geo party members have their CRIT DMG increased by 40%. Additionally, when Lumi uses Million Ton Crush in her Ultimate Power Form, the CRIT DMG of that attack increases by an additional 150%. 

Moonsign: Ascendant Gleam: When Lumi uses Heavy Overdrive Hammer in her Super Power Form, or when she uses Million Ton Crush in her Ultimate Power Form, an instance of Moondrift Harmony will be triggered. For this instance of Moondrift Harmony, all Hydro and Geo characters in the party will be considered to have applied their Elements in the reaction.

---

## Constellation C4 - Expert Instinct

Within 5s after Moondrift Harmony is triggered, increase the DEF of Linnea and your current active character by 25% respectively. When Linnea is on field, the DEF-increasing effect can be stacked.

---

## Constellation C6 - Golden Beagle's Dream

The effects of the Constellation, Provisional Classification, are enhanced: When unleashing Elemental Skill Countermeasure: Lumi's Battle Cry!, or when triggering Moondrift Harmony, Linnea will directly gain the maximum number of Field Catalog stacks. When consuming Field Catalog, consume twice the original number of stacks, such that the increase in DMG will be boosted to 150% of the original value.

Moonsign: Ascendant Gleam: Lunar-Crystallize Reaction DMG dealt by nearby party members is elevated by 25%.
> **実装ノート:** ElevationMod.