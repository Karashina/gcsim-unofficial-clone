package glog

import (
	"regexp"
	"strings"
)

// translations provides Japanese translations for English log messages
var translations = map[string]string{
	// Character abilities and events
	"Action extending timed Nightsoul Blessing":     "夜魂の加護の時間延長アクション",
	"Alhaitham left the field, mirror lost":         "アルハイゼンが場を離れ、鏡が失われました",
	"Alley gained stack":                             "路地でスタックを獲得",
	"Alley lost stack":                               "路地でスタックを失いました",
	"Arlecchino C6 dmg add":                          "アルレッキーノC6ダメージ追加",
	
	// Game mechanics
	"Bell activated":                                 "ベルが発動",
	"Bell ticking":                                   "ベルが鳴動中",
	"Blood Blossom checking for tick":               "血の花のティック確認中",
	"Blood Blossom ticked":                           "血の花がティック",
	"Blood Debt Directive checking for tick":        "血債指令のティック確認中",
	"Bond of Life changed":                           "生命の契約が変更されました",
	
	// Constellation effects
	"C2 bonus dmg% applied":                          "C2ボーナスダメージ%適用",
	"C4: Spawned 1 droplet":                         "C4: 水滴を1個生成",
	"C6: Picked up 1 droplet":                       "C6: 水滴を1個収集",
	
	// Weapon effects
	"Cinnabar Spindle proc dmg add":                  "辰砂往生録の発動ダメージ追加",
	"Citlali C1 proc dmg add":                        "シトラリC1発動ダメージ追加",
	
	// Item and resource management
	"Crystal Shrapnel gained from Burst":            "元素爆発から岩晶の欠片を獲得",
	"Crystal Shrapnel gained from Crystallise":      "結晶反応から岩晶の欠片を獲得",
	"Did not find any Bunnies":                      "ウサギが見つかりませんでした",
	"Directive upgraded":                             "指令がアップグレードされました",
	"DreamDrifter effect cancelled":                  "夢遊効果がキャンセルされました",
	"Energy gained from Crystallise":                "結晶反応から元素エネルギーを獲得",
	
	// Character specific effects
	"Escoffier C2 proc dmg add":                      "エスコフィエC2発動ダメージ追加",
	"Foul Legacy activated":                          "魔王武装が発動",
	"Gained Fanfare":                                 "喝采を獲得",
	
	// Weapon stacks
	"Haran gained a wavespike stack":                 "波乱月白経津がスタックを獲得",
	"Husk check for off-field stack":               "華館の殻スタックをオフフィールドで確認",
	"Husk gained off-field stack":                   "華館の殻がオフフィールドでスタック獲得",
	"Husk gained on-field stack":                    "華館の殻がオンフィールドでスタック獲得", 
	"Husk lost stack":                               "華館の殻がスタックを失いました",
	
	// Lynette specific
	"Lynette Bogglecat Box added":                   "リネットのにゃんこ箱が追加されました",
	
	// Lyney specific
	"Lyney C2 stack added":                          "リネC2スタックが追加されました",
	"Lyney C2 started":                              "リネC2が開始されました",
	"Lyney Grin-Malkin Hat added":                   "リネの笑顔の猫帽子が追加されました",
	"Lyney Grin-Malkin Hat removed":                 "リネの笑顔の猫帽子が削除されました",
	"Lyney Prop Surplus stack added":               "リネの道具余剰スタックが追加されました",
	"Lyney Prop Surplus stacks removed":            "リネの道具余剰スタックが削除されました",
	
	// General game mechanics
	"Mirror count is 0, omitting reduction":         "鏡の数が0のため、減少を省略",
	
	// Oz (Fischl's elemental skill)
	"Oz activated":                                  "オズが発動",
	"Oz not removed, src changed":                  "オズは削除されません、ソースが変更されました",
	"Oz removed":                                    "オズが削除されました",
	"Oz ticked":                                     "オズがティック",
	"Recasting oz":                                  "オズを再詠唱",
	
	// Character buffs and status
	"Paramita status extension for burst":          "波羅蜜ステータスの元素爆発延長",
	"Picked up snack":                               "おやつを拾いました",
	
	// Qiqi specific
	"Qiqi C1 Activation - Adding 2 energy":         "七七C1発動 - エネルギーを2追加",
	
	// Weapon procs
	"Redhorn proc dmg add":                          "赤角石塵滅砕の発動ダメージ追加",
	
	// Rosaria specific
	"Rosaria A1 activation":                         "ロサリアA1発動",
	"Rosaria A4 activation":                         "ロサリアA4発動",
	
	// Sethos specific
	"Sethos A4 proc dmg add":                        "セトスA4発動ダメージ追加",
	
	// Additional character specific abilities
	"Shenhe Quill proc dmg add":                     "申鶴の羽織発動ダメージ追加",
	"Sigewinne A1 proc dmg add":                     "シグウィンA1発動ダメージ追加",
	"Skill Tick Debug":                              "スキルティックデバッグ",
	"Skill: Spawned 3 droplets":                    "スキル: 水滴を3個生成",
	"Snack exploded by itself":                     "おやつが自爆しました",
	"Snack spawned":                                "おやつが生成されました",
	
	// Zhongli stele
	"Stele checking for tick":                      "岩柱のティック確認中",
	"Stele ticked":                                 "岩柱がティック",
	
	// Fontaine characters
	"Summoned Salon Solitaire":                     "サロン・ソリテールを召喚",
	"Summoned Singer of Many Waters":               "多くの水の歌手を召喚",
	
	// Weapon stacks continued
	"Surf's Up gained stack":                       "サーフズアップがスタック獲得",
	"Surf's Up lost stack":                         "サーフズアップがスタック失失",
	"Triggered Xiangling C2 explosion":             "香菱C2爆発をトリガー",
	"Tulaytullah's Remembrance gained stack via timer": "トゥライトゥラの記憶がタイマーでスタック獲得",
	
	// Kachina specific
	"Twirly checking for Ride attack":              "トゥイリーのライド攻撃を確認中",
	"Twirly checking for tick":                     "トゥイリーのティック確認中", 
	"Twirly ticked":                                "トゥイリーがティック",
	
	// Other mechanics
	"Valesa Tackle Hit":                            "ヴァレサタックルヒット",
	"Vermillion stack gained":                      "朱砂スタック獲得",
	"Xianyun A4 proc dmg add":                      "閑雲A4発動ダメージ追加",
	"Xianyun Adeptal Assistance stack consumed":   "閑雲の仙助スタック消費",
	"Xiao C6 activated":                            "魈C6発動",
	
	// Yaoyao mechanics
	"Yuegui (Jumping) removed":                     "月桂（ジャンプ）削除",
	"Yuegui (Jumping) summoned":                    "月桂（ジャンプ）召喚",
	"Yuegui (Throwing) removed":                    "月桂（投擲）削除", 
	"Yuegui (Throwing) summoned":                   "月桂（投擲）召喚",
	
	// Yun Jin specific
	"Yun Jin Party Elemental Types (A4)":          "雲菫パーティ元素タイプ（A4）",
	
	// Ascension effects
	"a1 adding crit rate":                          "A1会心率追加",
	"a1 adding flat dmg":                           "A1固定ダメージ追加",
	"a1 dash pp slide":                             "A1ダッシュPPスライド",
	"a1 infusion added":                            "A1元素付与追加",
	"a1 tailoring triggered":                       "A1仕立てトリガー",
	"a1 tapestry triggered":                        "A1タペストリートリガー",
	"a4 energy restore stacks":                     "A4エネルギー回復スタック",
	"a4 gained stack":                              "A4スタック獲得",
	"a4 triggered":                                 "A4トリガー",
	"add detector stack":                           "検出器スタック追加",
	
	// General additions
	"adding a1":                                    "A1追加",
	"adding c1":                                    "C1追加",
	"adding c2":                                    "C2追加", 
	"adding c6":                                    "C6追加",
	"adding delay due to falling":                  "落下による遅延追加",
	"adding energy":                                "エネルギー追加",
	"adding nilou a4 bonus":                        "ニィロウA4ボーナス追加",
	"adding shrapnel buffs":                        "欠片バフ追加",
	"adding star jade":                             "星玉追加",
	"adding stars":                                 "星追加",
	
	// Amber specific
	"amber exploding bunny":                        "アンバーの爆弾ウサギ",
	
	// Weapon procs continued
	"amenoma proc'd":                               "天目流発動",
	"applying cooldown modifier":                   "クールダウン修正子を適用",
	"archaic petra proc'd":                         "悠久の磐岩発動",
	
	// Klee specific
	"attempted klee skill cancel without burst":   "元素爆発なしでクレーのスキルキャンセルを試行",
	
	// Ayato specific
	"ayato a1 proc'd":                              "綾人A1発動",
	"ayato a1 set namisen stacks to max":          "綾人A1が波閃スタックを最大に設定",
	"ayato c6 proc'd":                              "綾人C6発動",
	
	// Barbara specific
	"barbara heal and wet ticking":                 "バーバラの回復と水元素付与ティック",
	"barbara melody loop ticking":                  "バーバラのメロディループティック",
	"barbara skill extended from a4":              "バーバラのスキルがA4から延長",
	
	// Beidou specific
	"beidou Q (active) on icd":                     "北斗の元素爆発（アクティブ）がICD中",
	"beidou Q proc'd":                              "北斗の元素爆発発動",
	
	// Bennett specific
	"bennett field - adding attack":                "ベネットのフィールド - 攻撃力追加",
	"bennett field ticking":                        "ベネットのフィールドティック",
	
	// General combat mechanics
	"blind spot entered":                           "死角に侵入",
	"bonus":                                        "ボーナス",
	"bounce: adding a1 stack from c1":             "バウンス: C1からA1スタック追加",
	"breakthrough state added":                     "突破状態追加",
	"breakthrough state deleted":                   "突破状態削除",
	"burst activated":                              "元素爆発発動",
	
	// Constellation mechanics
	"c1 reducing skill cooldown":                   "C1スキルクールダウン減少",
	"c1 restoring energy":                          "C1エネルギー回復",
	"c1 spawning rock doll":                        "C1岩人形生成",
	"c1: skill duration is extended":              "C1: スキル持続時間延長",
	"c2 activated":                                 "C2発動",
	"c2_stacks":                                    "C2スタック",
	"c4 activated":                                 "C4発動",
	"c4 proc'd on attack":                          "C4攻撃時発動",
	"c4 spawning kinu":                             "C4キヌ生成",
	"c4 stacks set to 0":                          "C4スタックを0に設定",
	"c4 triggered on damage":                       "C4ダメージ時トリガー",
	"c6 - adding star jade":                        "C6 - 星玉追加",
	"c6 adding crit DMG":                           "C6会心ダメージ追加",
	"c6 adding crit dmg":                           "C6会心ダメージ追加",
	"c6 buff extended":                             "C6バフ延長",
	"c6 expiry on":                                 "C6期限切れオン",
	"c6 proc dmg add":                              "C6発動ダメージ追加",
	
	// Status effects
	"calamity buff expired":                        "災厄バフ期限切れ",
	"calamity buff expiry check ignored, src diff": "災厄バフ期限切れチェック無視、ソース異なる",
	"calamity gained stack":                        "災厄スタック獲得",
	"cd":                                           "CD",
	"char":                                         "キャラ",
	
	// Charlotte specific
	"charlotte c4 adding dmg%":                     "シャルロットC4ダメージ%追加",
	
	// Chongyun specific  
	"chongyun adding infusion on swap":             "重雲スワップ時元素付与追加",
	"chongyun adding infusion":                     "重雲元素付与追加",
	"chongyun c4 recovering 2 energy":             "重雲C4エネルギー2回復",
	
	// Nightsoul mechanics
	"clear nightsoul points":                       "夜魂ポイントクリア",
	
	// Clorinde specific
	"clorinde healing surpressed":                  "クロリンデ回復抑制",
	
	// Coil mechanics
	"coil stack gained":                            "コイルスタック獲得",
	
	// Collei specific
	"collei a1 proc":                               "コレイA1発動",
	"collei a1 tick ignored, src diff":            "コレイA1ティック無視、ソース異なる",
	"collei a4 proc":                               "コレイA4発動",
	"collei c2 proc":                               "コレイC2発動",
	
	// Construct mechanics
	"construct spawning rock doll":                 "岩造物が岩人形を生成",
	"consume nightsoul points":                     "夜魂ポイント消費",
	"cooldown":                                     "クールダウン",
	"cr":                                           "会心率",
	
	// Weapon mechanics
	"crescent pike active":                         "流月の針アクティブ",
	"crimson witch 4pc adding stack":              "燃え盛る炎の魔女4セットスタック追加",
	
	// Combat mechanics
	"damage counter reset":                         "ダメージカウンターリセット",
	"damage reset timer set":                       "ダメージリセットタイマー設定",
	"damaging marked target":                       "マーク対象へのダメージ",
	
	// Dash mechanics
	"dash cd hitlag extended":                      "ダッシュCDヒットラグ延長",
	"dash cooldown triggered":                      "ダッシュクールダウントリガー",
	"dash lockout evaluation hitlag extended":     "ダッシュロックアウト評価ヒットラグ延長",
	"dash lockout evaluation started":             "ダッシュロックアウト評価開始",
	"dash on cooldown":                             "ダッシュクールダウン中",
	
	// Artifact and weapon stacks
	"declension stack gained":                      "デクレンションスタック獲得",
	"deep galleries 4pc stop playing":             "深林の記憶4セット演奏停止",
	
	// Target mechanics
	"default target changed on enemy death":       "敵の死亡でデフォルトターゲット変更",
	"default target is dead":                      "デフォルトターゲットが死亡",
	"defenderswill-4pc not implemented":           "守護の心4セット未実装",
	
	// Dehya specific
	"dehya a1 reducing redmane's blood dmg":       "ディシアA1が烈毛の血ダメージ軽減",
	"dehya mitigating dmg":                         "ディシアがダメージ軽減",
	"dehya-sanctum-c2-damage activated":           "ディシア聖所C2ダメージ発動",
	
	// Diona specific  
	"diona c6 incomming heal bonus activated":     "ディオナC6受信回復ボーナス発動",
	
	// Artifact sets
	"dm 4pc proc":                                  "DM4セット発動",
	"dmc-c4-triggered":                             "DMCC4トリガー",
	"dmg%":                                         "ダメージ%",
	
	// Weapon effects
	"dockhands-assistant adding stack":            "船渠の助手スタック追加",
	"doll attacking":                               "人形が攻撃",
	
	// Dori specific
	"dori a1 proc":                                 "ドリーA1発動",
	
	// Energy mechanics
	"draining energy":                              "エネルギー消耗",
	
	// Elemental reactions
	"ec expired":                                   "感電期限切れ",
	"ec wane":                                      "感電減衰",
	
	// Artifact effects
	"echoes 4pc adding dmg":                        "追憶4セットダメージ追加",
	"echoes 4pc failed to proc due icd":          "追憶4セットICD理由で発動失敗",
	"echoes 4pc failed to proc due to chance":    "追憶4セット確率理由で発動失敗",
	
	// Elemental application
	"ele app counter reset":                        "元素付与カウンターリセット",
	"ele app reset timer set":                     "元素付与リセットタイマー設定",
	"ele icd check":                                "元素ICDチェック",
	"ele lookup ok":                                "元素ルックアップOK",
	"em snapshot":                                  "元素熟知スナップショット",
	
	// Emilie specific
	"emilie c1 proc'd":                             "エミリーC1発動",
	"endoftheline proc":                            "エンドオブライン発動",
	
	// Enemy mechanics
	"enemy dead":                                   "敵死亡",
	"enemy hitlag - extending mods":               "敵ヒットラグ - 修正子延長",
	"enemy skipping tick":                          "敵ティックスキップ",
	
	// Weapon snapshots
	"engulfing lightning snapshot":                 "草薙の稲光スナップショット",
	
	// Nightsoul mechanics
	"enter nightsoul blessing":                     "夜魂の加護に入る",
	"exit nightsoul blessing":                      "夜魂の加護から出る",
	
	// Eula specific
	"eula a4 reset skill cd":                       "優菈A4スキルCD リセット",
	"eula burst add stack":                         "優菈元素爆発スタック追加",
	"eula burst started":                           "優菈元素爆発開始",
	"eula burst triggering":                        "優菈元素爆発トリガー",
	"eula c6 add additional stack":                 "優菈C6追加スタック追加",
	"eula: grimheart stack":                        "優菈: 氷魂スタック",
	
	// Action execution
	"executed noop wait(0)":                        "無操作wait(0)実行",
	"executed swap":                                "スワップ実行",
	"executed wait":                                "待機実行",
	
	// Status effects
	"expiry (without hitlag)":                      "期限切れ（ヒットラグなし）",
	"expiry":                                       "期限切れ",
	
	// Weapon and artifact effects continued
	"fading twillight cycle changed":               "薄暮の実サイクル変更",
	"faruzan a4 proc dmg add":                      "ファルザンA4発動ダメージ追加",
	"favonius proc'd":                              "西風発動",
	"flower of paradise lost 4pc adding stack":    "楽園の絶花4セットスタック追加",
	"flower-wreathed feathers cleared":             "花飾りの羽根クリア",
	"flower-wreathed feathers proc'd":              "花飾りの羽根発動",
	"foliarincision proc dmg add":                  "フォリアインシジョン発動ダメージ追加",
	"forestregalia leaf ignored":                   "森林のレガリア葉無視",
	"forestregalia proc'd":                         "森林のレガリア発動",
	"freedomsworn gained sigil":                    "蒼古なる自由への咆哮印章獲得",
	
	// Freminet specific
	"freminet a4 proc":                             "フレミネA4発動",
	"freminet c4 proc":                             "フレミネC4発動", 
	"freminet c6 proc":                             "フレミネC6発動",
	"freminet skill stacks gained":                 "フレミネスキルスタック獲得",
	
	// Weapon stacks continued
	"fruitoffulfillment gained stack":              "満願の実スタック獲得",
	"fruitoffulfillment lost stack":                "満願の実スタック失失",
	"fruitoffulfillment stack loss check ignored, src diff": "満願の実スタック失失チェック無視、ソース異なる",
	
	// Character specific effects continued
	"gained Gracious Rebuke from C1 N5":           "C1N5から恩恵の叱責獲得",
	"gained namisen stack":                         "波閃スタック獲得",
	"gambler-4pc proc'd":                           "ギャンブラー4セット発動",
	"generate nightsoul points":                    "夜魂ポイント生成",
	"geo-traveler field ticking":                   "岩主人公フィールドティック",
	"gilded dreams proc'd":                         "金メッキの夢4セット発動",
	"golden troupe 4pc lost":                       "黄金の劇団4セット失失",
	"golden troupe 4pc proc'd":                     "黄金の劇団4セット発動",
	
	// Guoba (Xiangling) specific
	"guoba hit by faruzan pressurized collapse":   "グゥオバがファルザンの圧縮崩壊に命中",
	"guoba hit by sucrose E":                       "グゥオバがスクロースEに命中",
	"guoba self infusion applied":                  "グゥオバ自己元素付与適用",
	
	// Weapon effects continued
	"hakushin proc'd":                              "白辰の輪発動",
	"heartstrings update stacks":                   "ハートストリングススタック更新",
	
	// Heizou specific
	"heizou a4 triggered":                          "鹿野院平蔵A4トリガー",
	"heizou-c6 adding stats":                       "鹿野院平蔵C6ステータス追加",
	
	// General mechanics
	"hp_drained":                                   "HP消耗",
	"hunterspath proc dmg add":                     "ハンターズパス発動ダメージ追加",
	"husk stack loss check ignored, src diff":     "華館の殻スタック失失チェック無視、ソース異なる",
	"icd (without hitlag)":                         "ICD（ヒットラグなし）",
	"index":                                        "インデックス",
	"infusion added":                               "元素付与追加",
	
	// Stamina mechanics
	"insufficient stam: charge attack":             "スタミナ不足: 重撃",
	"insufficient stam: dash":                      "スタミナ不足: ダッシュ",
	
	// Itto specific
	"itto burst":                                   "荒瀧一斗元素爆発",
	"itto-a1 atkspd stacks increased":              "荒瀧一斗A1攻撃速度スタック増加",
	"itto-a1 reset atkspd stacks":                  "荒瀧一斗A1攻撃速度スタックリセット",
	"itto-a4 applied":                              "荒瀧一斗A4適用",
	
	// Jean specific
	"jean c1 adding 40% dmg":                       "ジン C1 40%ダメージ追加",
	"jean self swirling":                           "ジン自己拡散",
	"jean-c6 not implemented":                      "ジンC6未実装",
	
	// Kaeya specific
	"kaeya a4 proc":                                "ガイアA4発動",
	"kaeya burst tick ignored, src diff":          "ガイア元素爆発ティック無視、ソース異なる",
	"kaeya-c2 proc'd":                              "ガイアC2発動",
	
	// Kazuha specific
	"kazuha a4 proc":                               "楓原万葉A4発動",
	"kazuha q src check ignored, src diff":        "楓原万葉元素爆発ソースチェック無視、ソース異なる",
	"kazuha-c2 ticking":                            "楓原万葉C2ティック",
	
	// Keqing specific
	"keqing c2 proc'd":                             "刻晴C2発動",
	
	// Kachina mechanics
	"kinu killed on attack":                        "キヌが攻撃で倒れました",
	"kinu spawned":                                 "キヌが生成されました",
	
	// Kokomi specific
	"kokomi c2 proc'd":                             "心海C2発動",
	
	// Artifact sets
	"lavawalker 2 pc not implemented":             "溶岩流浪者2セット未実装",
	
	// Layla specific
	"layla c4 adding damage":                       "レイラC4ダメージ追加",
	
	// Elemental reactions continued
	"lc expired":                                   "レザー期限切れ",
	"lc wane":                                      "レザー減衰",
	
	// Weapon effects
	"lostprayer gained stack":                      "四風原典スタック獲得",
	"maiden 4pc proc":                              "愛される少女4セット発動",
	"mappa-mare adding stack":                      "万国諸海の図スタック追加",
	
	// Character mechanics
	"marked by Lifeline":                           "生命線でマーク",
	"mirror decrease ignored, src diff":           "鏡減少無視、ソース異なる",
	"mirror overflowed":                            "鏡オーバーフロー",
	
	// Mizuki specific
	"mizuki c1 proc":                               "ミズキC1発動",
	
	// General mechanics
	"mod extended":                                 "修正子延長",
	
	// Mona specific
	"mona bubble on target":                        "モナのバブルがターゲットに",
	"mona-a1 phantom added":                        "モナA1幻影追加",
	
	// Weapon effects continued
	"moonglow add damage":                          "不滅の月華ダメージ追加",
	"moonpiercer leaf ignored":                     "穿月の矢葉無視",
	"moonpiercer proc'd":                           "穿月の矢発動",
	
	// Nahida specific
	"nahida c2 buff":                               "ナヒーダC2バフ",
	"namisen add damage":                           "波閃ダメージ追加",
	
	// Nightsoul mechanics
	"nightsoul ended, falling":                     "夜魂終了、落下中",
	
	// Artifact effects
	"noblesse 4pc proc":                            "旧貴族のしつけ4セット発動",
	
	// Noelle specific
	"noelle burst":                                 "ノエル元素爆発",
	"noelle c6 extension applied":                  "ノエルC6延長適用",
	
	// Ocean-Hued Clam set
	"ohc bubble accumulation":                      "海染硨磲バブル蓄積",
	"ohc bubble activated":                         "海染硨磲バブル発動",
	
	// Artifact sets continued
	"paleflame gained stack":                       "蒼白の炎4セットスタック獲得",
	
	// Particle mechanics
	"particle hp threshold triggered":              "粒子HPしきい値トリガー",
	"performing CA":                                "重撃実行",
	
	// Weapon stacks
	"portable-power-saw adding stack":              "ポータブルパワーソースタック追加",
	"pp slide activated":                           "PPスライド発動",
	"prop_surplus_stacks":                          "道具余剰スタック",
	"prospectors-drill adding stack":               "採掘者の鋼鑽スタック追加",
	
	// Raiden specific
	"raiden c6 triggered":                          "雷電将軍C6トリガー",
	"random energy on normal":                      "通常攻撃でランダムエネルギー",
	"range-gauge adding stack":                     "レンジゲージスタック追加",
	"remove_reason":                                "削除理由",
	"resolve stacks gained":                        "願力スタック獲得",
	"resolve stacks":                               "願力スタック",
	
	// Royal weapon series
	"royal stacked":                                "ロイヤルスタック",
	
	// Weapon effects continued  
	"sacrificial jade gained buffs":                "祭礼の玉璋バフ獲得",
	"sacrificial jade lost buffs":                  "祭礼の玉璋バフ失失",
	"sacrificial proc'd":                           "祭礼発動",
	
	// Character mechanics continued
	"sanctum added":                                "聖域追加",
	"sanctum picked up":                            "聖域拾得",
	"sapwood leaf ignored":                         "サップウッド葉無視",
	"sapwood proc'd":                               "サップウッド発動",
	
	// Sara specific
	"sara attack buff applied":                     "九条裟羅攻撃バフ適用",
	
	// Sayu specific
	"sayu c2 adding 3.3% dmg":                      "早柚C2 3.3%ダメージ追加",
	
	// Weapon stacks continued
	"scarletsands adding stack":                    "赤砂の杖スタック追加",
	"scarletsands did not gain stacks due to icd": "赤砂の杖ICDによりスタック獲得せず",
	"scarletsands icd counter decreased":           "赤砂の杖ICDカウンター減少",
	"scarletsands icd counter increased":           "赤砂の杖ICDカウンター増加",
	
	// Game mechanics
	"scent generated":                              "香り生成",
	"scent reset":                                  "香りリセット",
	"self reaction occured":                        "自己反応発生",
	"serpents subtlety generated":                  "蛇影の微妙さ生成",
	"serpents subtlety reduced":                    "蛇影の微妙さ減少",
	"set target direction to closest enemy":       "最も近い敵にターゲット方向設定",
	"set target direction":                         "ターゲット方向設定",
	
	// Shenhe specific
	"shenhe-c4 adding dmg bonus":                   "申鶴C4ダメージボーナス追加",
	"shenhe-c4 stack gained":                       "申鶴C4スタック獲得",
	
	// Shield mechanics
	"shield added":                                 "シールド追加",
	"shield bonus added":                           "シールドボーナス追加",
	"shield expired":                               "シールド期限切れ",
	"shield overridden":                            "シールド上書き",
	
	// General mechanics continued
	"shrapnel":                                     "欠片",
	"skill mult applied":                           "スキル倍率適用",
	
	// Yae Miko specific
	"sky kitsune thunderbolt":                      "天狐雷鳴",
	"sky kitsune tick at level":                    "天狐レベルティック",
	
	// Weapon effects
	"skywardatlas proc'd":                          "天空の書発動",
	"sodp 4pc adding dmg":                          "沈淪の心4セットダメージ追加",
	"spawning doll":                                "人形生成",
	
	// Debug info
	"src":                                          "ソース",
	"stack":                                        "スタック", 
	"stacks":                                       "スタック",
	"stam mod added":                               "スタミナ修正子追加",
	"starting hp set":                              "開始HP設定",
	
	// Status effects
	"status added":                                 "ステータス追加",
	"status extended":                              "ステータス延長",
	"status refreshed":                             "ステータス更新",
	"sturdy bone buff":                             "頑丈な骨バフ",
	
	// Sucrose specific
	"sucrose a1 triggered":                         "スクロースA1トリガー",
	"sucrose a4 triggered":                         "スクロースA4トリガー",
	"sucrose c4 reducing E CD":                     "スクロースC4元素スキルCD減少",
	
	// State mechanics
	"switching to bike state":                      "バイク状態に切り替え",
	"switching to ring state":                      "リング状態に切り替え",
	"swordofnarzissenkreuz arkhe":                  "ナルツィセンクロイツの剣アルケー",
	
	// Target mechanics
	"target position changed":                      "ターゲット位置変更",
	
	// Tartaglia specific
	"tartaglia c4 applied":                         "タルタリヤC4適用",
	"tartaglia c4 src check ignored, src diff":    "タルタリヤC4ソースチェック無視、ソース異なる",
	
	// Artifact sets continued
	"thunderfury 4pc proc":                         "雷のような怒り4セット発動",
	"thundersoother 2 pc not implemented":         "雷を鎮める尊者2セット未実装",
	"tom 4pc proc":                                 "千岩牢固4セット発動",
	
	// Traveler Electro specific
	"travelerelectro Q (active) on icd":           "旅人雷元素爆発（アクティブ）ICD中",
	"travelerelectro Q proc'd":                     "旅人雷元素爆発発動",
	"travelerelectro abundance amulet generated":   "旅人雷豊穣のお守り生成",
	
	// Traveler Hydro specific
	"travelerhydro a4 adding dmg bonus":           "旅人水A4ダメージボーナス追加",
	
	// General mechanics
	"tri-karma cd reduced":                         "三業CD減少",
	"ttds activated":                               "龍殺しの英傑譚発動",
	"update shield hp":                             "シールドHP更新",
	"update shield level":                          "シールドレベル更新",
	"using temporary target direction":             "一時的ターゲット方向使用",
	
	// Weapon effects
	"verdict adding skill dmg":                     "裁決スキルダメージ追加",
	"verdict adding stack":                         "裁決スタック追加",
	
	// Wanderer specific
	"wanderer-a4 available":                        "放浪者A4利用可能",
	"wanderer-a4 proc'd":                           "放浪者A4発動",
	"wavebreaker dmg calc":                         "波乱月白経津ダメージ計算",
	"widsith proc'd":                               "流浪楽章発動",
	
	// Xiao specific
	"xiao burst damage bonus":                      "魈元素爆発ダメージボーナス",
	"xiao c6 active, Xiao E used, no charge used, no CD": "魈C6アクティブ、魈元素スキル使用、チャージ未使用、CDなし",
	
	// Xilonen specific
	"xilonen c4 proc dmg add":                      "シロネンC4発動ダメージ追加",
	
	// Yanfei specific
	"yanfei charge attack consumed seals":          "煙緋重撃で丹火の印消費",
	"yanfei gained a seal from normal attack":      "煙緋通常攻撃で丹火の印獲得",
	"yanfei gained max seals":                      "煙緋最大丹火の印獲得",
	"yanfei gained seal from burst":                "煙緋元素爆発で丹火の印獲得",
	
	// Yaoyao specific
	"yaoyao a1 triggered":                          "ヤオヤオA1トリガー",
	
	// Yelan specific
	"yelan burst on skill":                         "夜蘭スキル時元素爆発",
	
	// Yun Jin specific
	"yunjin burst adding damage":                   "雲菫元素爆発ダメージ追加",
	
	// Dynamic message patterns (with placeholders)
	"mona-c6 stack gain check ignored, src diff":   "モナC6スタック獲得チェック無視、ソース異なる",
	"mona-c6 stack gained":                         "モナC6スタック獲得",
	"mona-c6 stacks reset via charge attack":       "モナC6スタック重撃でリセット",
	"mona-c6 stacks reset via timer":               "モナC6スタックタイマーでリセット",
}

// translationPatterns defines regex patterns for dynamic messages
var translationPatterns = []struct {
	pattern     *regexp.Regexp
	replacement string
}{
	// Itto SSS stacks pattern
	{
		pattern:     regexp.MustCompile(`^itto (\d+) SSS stacks from (.+)$`),
		replacement: "荒瀧一斗 $1 SSSスタック from $2",
	},
	// Target hit pattern
	{
		pattern:     regexp.MustCompile(`^target (.+) hit (\d+) times$`),
		replacement: "ターゲット $1 を $2 回攻撃",
	},
	// Crystal shrapnel firing pattern
	{
		pattern:     regexp.MustCompile(`^firing (\d+) crystal shrapnel$`),
		replacement: "$1 個の岩晶の欠片を発射",
	},
	// Droplet spawning patterns
	{
		pattern:     regexp.MustCompile(`^Burst: Spawned (\d+) droplets$`),
		replacement: "元素爆発: 水滴を $1 個生成",
	},
	{
		pattern:     regexp.MustCompile(`^Picked up (\d+) droplets$`),
		replacement: "水滴を $1 個収集",
	},
	{
		pattern:     regexp.MustCompile(`^Skill: Spawned (\d+) droplets$`),
		replacement: "スキル: 水滴を $1 個生成",
	},
	// Mirror patterns  
	{
		pattern:     regexp.MustCompile(`^Consumed (\d+) mirror\(s\)$`),
		replacement: "鏡を $1 個消費",
	},
	{
		pattern:     regexp.MustCompile(`^Gained (\d+) mirror\(s\)$`),
		replacement: "鏡を $1 個獲得",
	},
	// Weapon hit patterns
	{
		pattern:     regexp.MustCompile(`^(.+) hit by (.+)$`),
		replacement: "$1 が $2 に命中",
	},
	// Element interaction patterns
	{
		pattern:     regexp.MustCompile(`^(.+) came into contact with (.+)$`),
		replacement: "$1 が $2 と接触",
	},
	// Adding hitlag pattern
	{
		pattern:     regexp.MustCompile(`^(.+) applying hitlag: (.+)$`),
		replacement: "$1 ヒットラグ適用: $2",
	},
	// General weapon/effect proc patterns
	{
		pattern:     regexp.MustCompile(`^(.+) gained (.+) via (.+)$`),
		replacement: "$1 が $3 経由で $2 を獲得",
	},
}

// Translate returns the Japanese translation of an English message if available,
// otherwise returns the original message
func Translate(msg string) string {
	// First try exact match
	if translation, exists := translations[msg]; exists {
		return translation
	}
	
	// Then try pattern matching for dynamic messages
	for _, pattern := range translationPatterns {
		if pattern.pattern.MatchString(msg) {
			return pattern.pattern.ReplaceAllString(msg, pattern.replacement)
		}
	}
	
	return msg
}