package hutao

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/enemy"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

const (
	c6ICDKey = "hutao-c6-icd"
)

func (c *char) c6() {
	c.c6buff = make([]float64, attributes.EndStatType)
	c.c6buff[attributes.CR] = 1
	// 被ダメージ時の6命ノ星座発動をチェック
	c.Core.Events.Subscribe(event.OnPlayerHPDrain, func(args ...interface{}) bool {
		di := args[0].(*info.DrainInfo)
		if di.Amount <= 0 {
			return false
		}
		c.checkc6(false)
		return false
	}, "hutao-c6")
	// 被ダメージに関係なくシミュレーション開始から2秒ごとに6命ノ星座発動をチェック
	c.Core.Tasks.Add(func() { c.checkc6(true) }, 1) // HP設定後にチェック開始
}

func (c *char) checkc6(check1HP bool) {
	// 被ダメージとc6 ICDに関係なく2秒ごとに6命ノ星座発動をチェック
	c.QueueCharTask(func() {
		c.checkc6(true)
	}, 120)
	// 6命ノ星座がICD中かチェック
	if c.StatusIsActive(c6ICDKey) {
		return
	}
	// HPが25%以下かチェック
	if c.CurrentHPRatio() > 0.25 {
		return
	}
	// 2秒チェック用にHPが2未満か確認
	if check1HP && c.CurrentHP() >= 2 {
		return
	}
	// 死亡している場合はHP1で復活
	if c.CurrentHPRatio() <= 0 {
		c.SetHPByAmount(1)
	}

	// 会心率00%に増加
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("hutao-c6", 600),
		AffectedStat: attributes.CR,
		Amount: func() ([]float64, bool) {
			return c.c6buff, true
		},
	})

	c.AddStatus(c6ICDKey, 3600, false)
}

// 胡桃自身が付与した血梅香の影響を受けている敵を撃破した時、
// パーティーの他の全味方（胡桃自身を除く）の
// 会心率12%が15秒間アップする。
func (c *char) c4() {
	c.c4buff = make([]float64, attributes.EndStatType)
	c.c4buff[attributes.CR] = 0.12
	c.Core.Events.Subscribe(event.OnTargetDied, func(args ...interface{}) bool {
		t, ok := args[0].(*enemy.Enemy)
		// 敵でなければ何もしない
		if !ok {
			return false
		}
		if !t.StatusIsActive(bbDebuff) {
			return false
		}
		for i, char := range c.Core.Player.Chars() {
			// 胡桃には適用されない
			if c.Index == i {
				continue
			}
			char.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag("hutao-c4", 900),
				AffectedStat: attributes.CR,
				Amount: func() ([]float64, bool) {
					return c.c4buff, true
				},
			})
		}

		return false
	}, "hutao-c4")
}
