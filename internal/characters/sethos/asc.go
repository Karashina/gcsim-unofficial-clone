package sethos

func (c *char) reseta4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4stacks = 4
}

func (c *char) removea4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.a4stacks = 0
}
