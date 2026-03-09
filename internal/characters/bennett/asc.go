package bennett

// 溢れる情熱のCDを20%短縮する。
func (c *char) a1(cd int) int {
	if c.Base.Ascension < 1 {
		return cd
	}
	return int(float64(cd) * 0.8)
}

// 素晴らしい旅の範囲内で、溢れる情熱は以下の効果を得る：
//
// - CDが50%短縮される。
func (c *char) a4CD(cd int) int {
	if c.Base.Ascension < 4 || !c.StatModIsActive(burstFieldKey) {
		return cd
	}
	return int(float64(cd) * 0.5)
}

// 素晴らしい旅の範囲内で、溢れる情熱は以下の効果を得る：
//
// - ベネットはチャージレベル2の効果で打ち上げられなくなる。
func (c *char) a4NoLaunch() bool {
	return c.Base.Ascension >= 4 && c.StatModIsActive(burstFieldKey)
}
