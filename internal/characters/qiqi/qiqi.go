package qiqi

import (
	tmpl "github.com/Karashina/gcsim-unofficial-clone/internal/template/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/model"
)

const (
	talismanKey    = "qiqi-talisman"
	talismanICDKey = "qiqi-talisman-icd"
)

func init() {
	core.RegisterCharFunc(keys.Qiqi, NewChar)
}

type char struct {
	*tmpl.Character
	skillLastUsed     int
	skillHealSnapshot combat.Snapshot // 被弾時回復と継続回復の両方がこれを使用するため必要
}

// TODO: 未実装 - 6命ノ星座（復活メカニクス、シムには不向き）
// 4命ノ星座 - 敵の攻撃力減少、このシムバージョンでは有用ではない
func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = 80
	c.NormalHitNum = normalHitNum
	c.BurstCon = 3
	c.SkillCon = 5

	c.skillLastUsed = 0

	w.Character = &c

	return nil
}

// ターゲットのセットが正しく初期化されていることを確認
func (c *char) Init() error {
	c.a1()
	c.talismanHealHook()
	c.onNACAHitHook()
	if c.Base.Cons >= 2 {
		c.c2()
	}
	return nil
}

// 現在のキャラクターステータス（全モディファイア適用済み）を使用して動的に回復量を計算するヘルパー関数
func (c *char) healDynamic(healScalePer, healScaleFlat []float64, talentLevel int) float64 {
	atk := c.TotalAtk()
	heal := healScaleFlat[talentLevel] + atk*healScalePer[talentLevel]
	return heal
}

// スナップショットインスタンスから回復量を計算するヘルパー関数
func (c *char) healSnapshot(d *combat.Snapshot, healScalePer, healScaleFlat []float64, talentLevel int) float64 {
	atk := d.Stats.TotalATK()
	heal := healScaleFlat[talentLevel] + atk*healScalePer[talentLevel]
	return heal
}

func (c *char) AnimationStartDelay(k model.AnimationDelayKey) int {
	if k == model.AnimationXingqiuN0StartDelay {
		return 7
	}
	return c.Character.AnimationStartDelay(k)
}
