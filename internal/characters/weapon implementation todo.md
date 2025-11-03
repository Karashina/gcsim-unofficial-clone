# 実装時の注意点
- pkg\simulation\imports_char_gen.goに追加されたキャラクターのフォルダのパスを追加すること
- pkg\shortcut\characters.goに追加されたキャラクターのエイリアスを作ること
- pkg\core\keys\keys_char_gen.goに追加されたキャラクターの項目を作ること
- ほかのキャラクターの例をもとにconfig.yml、data_gen.textproto、キャラクター名_gen.goを作成すること

# 実装するキャラクター
- aino
    ID: 90000603
    element: Hydro
    weapon_class: CLAYMORE

Energymax: 50
Skillcon: 5
Burstcon: 3


備考: ineffa/Lauma/flins/neferをベースに設計すること。

- Normal Attack - Bish-Bash-Bosh Repair

Normal Attack
Performs up to 3 consecutive strikes.

Charged Attack
実装不要。

Plunging Attack
Plunges from mid-air to strike the ground below, damaging opponents along the path and dealing AoE DMG upon impact.

倍率スケーリング: aino_gen.goに記載すること。
Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11
1-Hit DMG 	66.5%	71.9%	77.3%	85.1%	90.5%	96.7%	105.2%	113.7%	122.2%	131.5%	140.7%
2-Hit DMG 	66.2%	71.6%	77.0%	84.7%	90.1%	96.2%	104.7%	113.1%	121.6%	130.8%	140.1%
3-Hit DMG 	49.2%×2	53.2%×2	57.2%×2	63.0%×2	67.0%×2	71.5%×2	77.8%×2	84.1%×2	90.4%×2	97.3%×2	104.2%×2
Plunge DMG 	74.6%	80.7%	86.7%	95.4%	101.5%	108.4%	118.0%	127.5%	137.0%	147.4%	157.8%
Low Plunge DMG 	149%	161%	173%	191%	203%	217%	236%	255%	274%	295%	316%
High Plunge DMG  186%	201%	217%	238%	253%	271%	295%	318%	342%	368%	394%

備考: doriを参考に実装すること。

- Skill - Musecatcher
Aino deals Hydro DMG to single target(Stage 1).
Then, Aino deals AoE Hydro DMG to nearby opponents(Stage 2).

Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11	Lv.12	Lv.13	Lv.14
Stage 1 DMG 	65.6%	70.5%	75.4%	82.0%	86.9%	91.8%	98.4%	105.0%	111.5%	118.1%	124.6%	131.2%	139.4%	147.6%
Stage 2 DMG 	188.8%	203.0%	217.1%	236%	250.2%	264.3%	283.2%	302.1%	321.0%	339.8%	358.7%	377.6%	401.2%	424.8%

CD: 10s

- Burst - Precision Hydronic Cooler

Aino deploys the "Cool Your Jets Ducky" to establish a "Focused Hydronic Cooling Zone".
While active, the Cool Your Jets Ducky will periodically fire water balls at nearby opponents, dealing Hydro DMG.

Level 	Lv.1	Lv.2	Lv.3	Lv.4	Lv.5	Lv.6	Lv.7	Lv.8	Lv.9	Lv.10	Lv.11	Lv.12	Lv.13
Water Ball DMG 	20.1%	21.6%	23.1%	25.1%	26.6%	28.2%	30.2%	32.2%	34.2%	36.2%	38.2%	40.2%	42.7%
Duration 	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s	14.0s

CD: 13.5s

- Asc 1 - Modular Efficiency Protocol
When the Moonsign is Ascendant, Her Elemental Burst Precision Hydronic Cooler is enhanced:
The Cool Your Jets Ducky will fire water balls more frequently, and the water balls will deal AoE Hydro DMG over a larger area of effect.

- Asc 4 - Structured Power Booster
Aino's Elemental Burst FlatDMG is increased by 50% of her Elemental Mastery.

- Cons 1
After Aino uses her Elemental Skill or her Elemental Burst,
her Elemental Mastery will be increased by 80 for 15s.
Also, he Elemental Mastery of other nearby active party members will be increased by 80 for 15s.
The Elemental Mastery-increasing effects of this Constellation do not stack.

- Cons 2
If Aino is off-field while the Focused Hydronic Cooling Zone of her Elemental Burst Precision Hydronic Cooler is active,
when your active party member hits a nearby opponent with an attack, the Cool Your Jets Ducky will fire an additional water ball at that opponent, dealing AoE Hydro DMG.
It has no Mult but it's FlatDmg is sum of 25% of Aino's ATK and 100% of her Elemental Mastery.
AttackTag of this DMG is considered as AttackTagElementalBurst. This effect can be triggered once every 5s.

- Cons 4
When the Elemental Skill hits an opponent, it will restore 10 Elemental Energy for Aino. Energy can be restored to her in this manner once every 10s.

- Cons 6
For the next 15s after using the Elemental Burst, DMG from nearby active characters' Electro-Charged, Bloom, Lunar-Charged, and Lunar-Bloom reactions is increased by 15%.
When the Moonsign is Ascendant, DMG from the aforementioned reactions will be further increased by 20%.

# data_gen.textprotoの内容 (このまま使用すること)
id: 90000603
key: "aino"
rarity: QUALITY_PURPLE
body: BODY_LOLI
region: 12
element: Water
weapon_class: WEAPON_CLAYMORE
icon_name: "UI_AvatarIcon_Ambor"
stats: {
   base_hp: 939
   base_atk: 20
   base_def: 51
   hp_curve: GROW_CURVE_HP_S4
   atk_curve: GROW_CURVE_ATTACK_S5
   def_cruve: GROW_CURVE_HP_S5
   promo_data: {
      max_level: 20
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
      }
   }
   promo_data: {
      max_level: 40
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 701
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 38
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 15
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
      }
   }
   promo_data: {
      max_level: 50
add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 1199.0605
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 64.999
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
         value: 25.6575
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 24
      }
   }
   promo_data: {
      max_level: 60
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 1863.1879
      }
add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 101.0002
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
         value: 39.8685
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 48
      }
   }
   promo_data: {
      max_level: 70
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 2361.2484
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 127.9992
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
         value: 50.526
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 48
      }
   }
   promo_data: {
      max_level: 80
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 2859.3089
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 154.9982
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
         value: 61.1835
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 72
      }
   }
   promo_data: {
      max_level: 90
      add_props: {
         prop_type: FIGHT_PROP_BASE_HP
         value: 3357.4395
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_DEFENSE
         value: 182.001
      }
      add_props: {
         prop_type: FIGHT_PROP_BASE_ATTACK
         value: 71.8425
      }
      add_props: {
         prop_type: FIGHT_PROP_ELEMENT_MASTERY
         value: 96
      }
   }
}
skill_details: {
   skill: 10262
   burst: 10265
   attack: 10261
   burst_energy_cost: 50
}
name_text_hash_map: 1021947691
