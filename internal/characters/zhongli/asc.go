package zhongli

// 玉璋シールドがダメージを受けると、強化される:
//
// - 強化されたキャラクターはシールド強度が5%上昇する。
//
// - 最大5回まで重ね掛け可能。玉璋シールドが消えるまで持続する。
func (c *char) a1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Player.Shields.AddShieldBonusMod("zhongli-a1", -1, func() (float64, bool) {
		if c.Tags["shielded"] == 0 {
			return 0, false
		}
		return float64(c.Tags["a1"]) * 0.05, true
	})
}

// 鍾離は最大HPに基づく追加ダメージを与える:
//
// - 通常攻撃、重撃、落下攻撃のダメージがHP上限の1.39%分増加する。
func (c *char) a4Attacks() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return 0.0139 * c.MaxHP()
}

// 鍾離は最大HPに基づく追加ダメージを与える:
//
// - 地心の岩柱、共鳴、長押しダメージがHP上限の1.9%分増加する。
func (c *char) a4Skill() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return 0.019 * c.MaxHP()
}

// 鍾離は最大HPに基づく追加ダメージを与える:
//
// - 天星のダメージがHP上限の33%分増加する。
func (c *char) a4Burst() float64 {
	if c.Base.Ascension < 4 {
		return 0
	}
	return 0.33 * c.MaxHP()
}
