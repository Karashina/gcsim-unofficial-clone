package lauma

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

func init() {
	core.RegisterCharFunc(keys.Lauma, NewChar)
}

type char struct {
	*tmpl.Character
	skillSrc    int
	burstLBBuff float64
	verdantDew  int
	moonSong    int
	paleHymn    int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.SkillCon = 5
	c.BurstCon = 3

	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	// チーム初期化のためこのキャラクターをムーンサイン保持候補としてマーク
	c.AddStatus("moonsignKey", -1, false)
	c.setupPaleHymnEffects()
	c.a0()
	c.a4()                // A4 AddAttackModを初期化
	c.c6Init()            // 6凸 Ascendant Elevationボーナスを初期化
	c.verdantDewCheck()   // 翠露の監視を初期化
	c.applyResReduction() // 耐性低下の監視を初期化
	return nil
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 11
	}
	return c.Character.AnimationStartDelay(k)
}

// verdantDewCheckは開花イベントをサブスクライブし、パーティに翠露を付与する。
// 翠露は最大3で、獲得は充填期間後にキューに入る。
func (c *char) verdantDewCheck() {
	c.Core.Events.Subscribe(event.OnBloom, func(args ...interface{}) bool {
		if !c.StatusIsActive("LB-Key") {
			return false
		}
		duradd := c.StatusDuration("dewchargingkey")
		c.AddStatus("dewchargingkey", 150, true) // 2.5秒の充填期間
		c.QueueCharTask(func() {
			if c.verdantDew < 3 {
				c.verdantDew++
			}
		}, 150+duradd)
		return true
	}, "lauma-verdant-dew")
}
