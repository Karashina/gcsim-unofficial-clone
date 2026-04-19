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

<EN>
Normal Attack
Performs up to 3 consecutive shots with a bow.

Charged Attack
Performs a more precise Aimed Shot with increased DMG.
While aiming, stone crystals will accumulate on the arrowhead. A fully charged crystalline arrow will deal Geo DMG.

Plunging Attack
Fires off a shower of arrows in mid-air before falling and striking the ground, dealing AoE DMG upon impact.

<JP>
通常攻撃
最大3段の連続射撃を行う。

重撃
ダメージがより高く、より精確な狙い撃ちを発動する。
照準時、岩晶を矢先に凝集させ、岩晶に満ちた矢で敵に岩元素ダメージを与える。

落下攻撃
空中から矢の雨を放ち、凄まじいスピードで落下し地面に衝撃を与え、落下時に範囲ダメージを与える。
 
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

<EN>
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
There are 2s ICD on generation.

<JP>
一回押し
ルミはとっておき形態で出撃し、近くの敵を継続的に攻撃して岩元素範囲ダメージを与える。近くに月籠が存在する場合、ルミは近くの敵に月結晶反応ダメージと見なされる岩元素範囲ダメージを与える。

連打
いざ、本気を見せる時だ！元素スキル発動後、元素スキルまたは通常攻撃を連打すると、リンネアはルミにキラキラの宝石をたくさん食べさせ、お腹いっぱいになったルミはさいごのきりふだ形態で出撃し、近くの敵に月結晶反応ダメージと見なされる極めて強力な岩元素範囲ダメージを1回与え、がんばり形態に切り替わる。
また、元素スキルを連打している間は、リンネアの中断耐性がアップする。

ルミ・出撃形態
ルミには3種類の出撃形態がある。

とっておき形態
ルミはポコポコハンマーで近くの敵を攻撃し続け、各攻撃ごとに岩元素範囲ダメージを2回与える。
近くに月籠が存在する場合、ルミはポコポコハンマーとパワーハンマーを交互に使用して近くの敵を攻撃し続ける。パワーハンマーは、月結晶反応ダメージと見なされる岩元素範囲ダメージを与える。

さいごのきりふだ形態
ルミはさいごのきりふだ100万トンハンマーで近くの敵に月結晶反応ダメージと見なされる極めて強力な岩元素範囲ダメージを1回与え、がんばり形態に切り替わる。

がんばり形態
他の形態ほどスゴくはないが、それでもスゴイ。
ルミはより長い間隔を空けながらポコポコハンマーで近くの敵を攻撃し続ける。各攻撃ごとに岩元素範囲ダメージを2回与える。

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

<EN>
Linnea summons Lumi to strike in Super Power Form, healing nearby party members. For a short duration, she will continuously heal nearby active party members based on Linnea's DEF.
If Lumi is already on the field when Linnea unleashes Elemental Burst, Lumi's active duration will be reset instead and her Strike Form will not change.

<JP>
超うでききの冒険者も、適度に休まないと！リンネアのもと、ルミはとっておき形態で出撃して、近くのチーム内キャラクター全員のHPを回復する。さらにその後の一定時間、近くのフィールド上にいるキャラクターのHPを継続的に回復する。回復量はリンネアの防御力に基づく。
元素爆発の発動時、ルミがすでに出撃している場合、ルミの出撃形態は変わらずに継続時間がリセットされる。

### Talent Scaling
Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11	Lv.12	Lv.13
Initial Healing Amount 	160.0% DEF+770	172.0% DEF+847	184.0% DEF+930	200.0% DEF+1020	212.0% DEF+1117	224.0% DEF+1219	240.0% DEF+1328	256.0% DEF+1444	272.0% DEF+1566	288.0% DEF+1694	304.0% DEF+1829	320.0% DEF+1971	340.0% DEF+2118
Continuous Healing 	32.0% DEF+154	34.4% DEF+169	36.8% DEF+186	40.0% DEF+204	42.4% DEF+223	44.8% DEF+243	48.0% DEF+265	51.2% DEF+288	54.4% DEF+313	57.6% DEF+338	60.8% DEF+365	64.0% DEF+394	68.0% DEF+423
Healing Duration 	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s	12.0s
CD 	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s	15.0s
Energy Cost 	60	60	60	60	60	60	60	60	60	60	60	60	60


---

## Ascension Passive A1 - Field Observation Notes

<EN>
When Lumi is present on the field, the Geo RES of opponents near Lumi will decrease by 15%.

Moonsign: Ascendant Gleam: Linnea's Elemental Skill Countermeasure: Lumi's Battle Cry! and Elemental Burst Memo: Survival Guide in Extreme Conditions are enhanced. After summoning Lumi, the Geo RES of opponents near Lumi will be further decreased by 15%.

<JP>
ルミがフィールド上にいる時、ルミの近くにいる敵の岩元素耐性-15%。

月兆・満照：リンネアの元素スキルルミのやっふー作戦と元素爆発絶体絶命サバイバル備忘録が強化され、ルミをフィールド上に呼び出した後、ルミの近くにいる敵の岩元素耐性がさらに15%ダウンする。

---

## Ascension Passive A4 - Universal Naturalist Archive

<EN>
Linnea will increase the Elemental Mastery of certain characters in your party according to your current active character. Increase in Elemental Mastery is based on 5% of Linnea's DEF. If your current active character:
· Is a Moonsign character: Increase this character's Elemental Mastery.
· Is not a Moonsign character: Increase Linnea's own Elemental Mastery.

<JP>
フィールド上にいるチーム内の自身のキャラクターに基づき、リンネアはチーム内の特定キャラクターの元素熟知をリンネアの防御力の5%分アップさせる。キャラクターの性質に応じた効果は以下の通り。
月兆キャラクター：そのキャラクターの元素熟知をアップさせる。
月兆キャラクター以外：リンネア自身の元素熟知をアップさせる。

---

## Ascension Passive　A0: Moonsign Benediction: Habitat Survey

<EN>
When a party member triggers a Hydro Crystallize reaction, it will be converted into the Lunar-Crystallize reaction, with every 100 DEF that Linnea has increasing Lunar-Crystallize's Base DMG by 0.7%, up to a maximum of 14%.

Additionally, when Linnea is in the party, the party's Moonsign will increase by 1 level.

<JP>
チーム内のキャラクターが水元素結晶反応を起こすと、月結晶反応へと変わり、リンネアの防御力に基づいて、チーム内キャラクターの与える月結晶反応の基礎ダメージがアップする。防御力100につき、月結晶反応の基礎ダメージ+0.7%。この方法でアップできるダメージは最大14%まで。

また、リンネアがチームにいる時、チームの月兆レベルが1アップする。

> **実装ノート:** ZibaiのA0と同じ効果。

---

## Constellation C1 - Provisional Classification

<EN>
When unleashing Elemental Skill Countermeasure: Lumi's Battle Cry!, or when triggering Moondrift Harmony, Linnea gains 6 stacks of the Field Catalog effect for 10s. Max 18 stacks. When nearby party members deal Lunar-Crystallize Reaction DMG, consume 1 stack of Field Catalog to increase the DMG dealt. The increase in DMG is equal to 75% of Linnea's DEF.
Additionally, when Lumi uses Million Ton Crush in her Ultimate Power Form, Linnea can consume up to 5 stacks of Field Catalog. Each stack will increase the DMG dealt by 150% of Linnea's DEF.

<JP>
月籠の共鳴を発動した後の8秒間、チーム内の全ての水元素と岩元素キャラクターの会心ダメージ+40%。さらに、ルミがさいごのきりふだ形態で100万トンハンマーを使用した際の会心ダメージ+150%。

月兆・満照：ルミがとっておき形態でパワーハンマーを使用、またはさいごのきりふだ形態で100万トンハンマーを使用した時、月籠の共鳴を1回発動する。なお、その際の月籠の共鳴は、チーム内の全ての水元素と岩元素キャラクターが元素反応に関与したものとみなされる。

---

## Constellation C2 - Tidings of Joy and Sorrow

<EN>
Within 8s after triggering Moondrift Harmony, all Hydro and Geo party members have their CRIT DMG increased by 40%. Additionally, when Lumi uses Million Ton Crush in her Ultimate Power Form, the CRIT DMG of that attack increases by an additional 150%. 

Moonsign: Ascendant Gleam: When Lumi uses Heavy Overdrive Hammer in her Super Power Form, or when she uses Million Ton Crush in her Ultimate Power Form, an instance of Moondrift Harmony will be triggered. For this instance of Moondrift Harmony, all Hydro and Geo characters in the party will be considered to have applied their Elements in the reaction.

<JP>

---

## Constellation C4 - Expert Instinct

<EN>
Within 5s after Moondrift Harmony is triggered, increase the DEF of Linnea and your current active character by 25% respectively. When Linnea is on field, the DEF-increasing effect can be stacked.

<JP>
月籠の共鳴を発動した後の5秒間、リンネア及びフィールド上にいる自身のチーム内キャラクターの防御力+25%。リンネアがフィールド上にいる場合、上記の防御力アップ効果は重ね掛けされる。

---

## Constellation C6 - Golden Beagle's Dream

<EN>
The effects of the Constellation, Provisional Classification, are enhanced: When unleashing Elemental Skill Countermeasure: Lumi's Battle Cry!, or when triggering Moondrift Harmony, Linnea will directly gain the maximum number of Field Catalog stacks. When consuming Field Catalog, consume twice the original number of stacks, such that the increase in DMG will be boosted to 150% of the original value.

Moonsign: Ascendant Gleam: Lunar-Crystallize Reaction DMG dealt by nearby party members is elevated by 25%.

<JP>
命ノ星座「未分類のモノたち」の効果が強化され、元素スキルルミのやっふー作戦、または月籠の共鳴が発動した時、リンネアは最大層数の「文献調査」効果を獲得する。また、「文献調査」を消費する際、本来の2倍の層数を消費し、ダメージアップ効果が元の150%になる。

月兆・満照：付近にいるチーム内キャラクターが与える月結晶反応ダメージが25%向上する。

> **実装ノート:** ElevationMod.

## フレーム等
//------------------------NA------------------------//
<N1>
relase: 22
N1 -> N2: 35

GU: 1U
ICDタグ: 通常
ICDグループ: 標準

<N2>
relase: 13
N2 -> N3: 34

GU: 1U
ICDタグ: 通常
ICDグループ: 標準

<N3>
relase: 47
N3 -> N1: 85

GU: 1U
ICDタグ: 通常
ICDグループ: 標準

//------------------------E------------------------// 
<E>
tE(tap-E) -> D: 17
tE -> ルミ召喚: 34
tE -> ルミ初撃: 108
ポコポコハンマーヒット間: 21
ポコポコハンマー2hitmark -> パワーハンマーhitmark: 88
パワーハンマーhitmark -> ポコポコハンマーhitmark: 61
とっておきポコポコハンマー2hitmark -> とっておきポコポコハンマー1hitmark: 120
tE -> 終了: 1560

mE(mash-E) -> D: 97
mE -> hitmark: 111
100万トンハンマーhitmark -> ポコポコハンマーhitmark1: 132
がんばりポコポコハンマー2hitmark -> がんばりポコポコハンマー1hitmark: 300

GU: 1U (ポコポコハンマー)
ICDタグ: 元素スキル
ICDグループ: 標準

//------------------------Q------------------------// 
<Q>
Q -> D: 106
Q -> 回復: 96
Q -> 継続回復開始: 158
回復間隔: 60
回復回数: 12回

CT開始: 2
ｴﾈﾙｷﾞｰ: 4

GU: 1U
ICDタグ: 元素爆発
ICDグループ: 無し

## 能力の詳細解説
ざっくりまとめると

ダメージ出力はほぼ100%スキル（ルミ）が担う
元素爆発は回復のみ、ダメージへの貢献なし
通常攻撃は切替CT待ちで振る程度
通常攻撃以外は一貫して防御力スケールなので、ビルド時に通常攻撃は考慮不要

元素スキル（ルミのやっふー作戦）

単押しと連打で形態が異なる
連打の猶予は13フレーム前後と割と短い

単押し（とっておき形態）

約2秒間隔でポコポコハンマーの2hit攻撃
月籠あり時はポコポコハンマーを2回行った後、パワーハンマーで月結晶ダメージ

連打（さいごのきりふだ→がんばり形態）

100万トンハンマーでパワーハンマー約4回分の月結晶ダメージを即時与える
だいたい1ローテ分ぐらいのダメージを一気に与える感じ
ただし、がんばり形態ではポコポコハンマーも減速するので、雑魚敵の一括処理やとどめで使うのがいいかも

その他

ルミは岩元素構造物ではない（鍾離共鳴・千織の召喚判定対象外）
粒子生成CTは約9秒、ポコポコハンマーのヒットに反応して生成

元素爆発（絶体絶命サバイバル備忘録）

60族、CT15秒、回復のみでダメージなし
回復量はそこまで多いわけではなく、ダメージのケアが主眼
フリーナのテンションは最大まで貯められる水準（検証済み）
ルミの再召喚が可能、召喚済みの場合は継続時間をリセット
ただし形態も引き継がれるため、がんばり形態中に爆発を使うと、がんばり形態のまま継続時間が延びてしまう点に注意

パッシブ天賦
A0（月兆の祝福・生息地研究）

茲白と同じ効果、DEF2000で上限（月結晶基礎ダメ+14%）
月兆レベル+1

A1（フィールド観察ノート）

満照かつルミが居ればOKという軽い条件で岩耐性-30%
満照でない場合は-15%

A4（博物分類図鑑）

月兆キャラへの熟知バフ、上限なし
DEF2000で熟知100付与