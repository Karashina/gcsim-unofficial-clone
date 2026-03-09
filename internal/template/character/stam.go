package character

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/glog"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player"
)

// ActionStam は重撃とダッシュのスタミナコストのデフォルト実装を提供する
// キャラクターはこれをオーバーライドすべき
func (c *Character) ActionStam(a action.Action, p map[string]int) float64 {
	switch a {
	case action.ActionCharge:
		// 20 片手剣（大半）
		// 25 長柄武器
		// 40/秒 両手剣
		// 50 法器
		switch c.Weapon.Class {
		case info.WeaponClassSword:
			return 20
		case info.WeaponClassSpear:
			return 25
		case info.WeaponClassCatalyst:
			return 50
		case info.WeaponClassClaymore:
			return 0
		case info.WeaponClassBow:
			return 0
		default:
			return 0
		}
	case action.ActionDash:
		// 1回あたり18
		return 18
	default:
		return 0
	}
}

func (c *Character) Dash(p map[string]int) (action.Info, error) {
	// ダッシュクールダウンのロジックを実行
	c.ApplyDashCD()

	// ダッシュ終了時にスタミナを消費
	c.QueueDashStaminaConsumption(p)

	length := c.DashLength()
	dashJumpLength := c.DashToJumpLength()
	return action.Info{
		Frames: func(a action.Action) int {
			switch a {
			case action.ActionJump:
				return dashJumpLength
			default:
				return length
			}
		},
		AnimationLength: length,
		CanQueueAfter:   dashJumpLength,
		State:           action.DashState,
	}, nil
}

// ダッシュCDを設定する。このダッシュ実行時にCDが残っていた場合、ダッシュをロックアウトする
func (c *Character) ApplyDashCD() {
	var evt glog.Event

	if c.Core.Player.DashCDExpirationFrame > c.Core.F {
		c.Core.Player.DashLockout = true
		c.Core.Player.DashCDExpirationFrame = c.Core.F + 1.5*60
		evt = c.Core.Log.NewEvent("dash cooldown triggered", glog.LogCooldownEvent, c.Index)
	} else {
		c.Core.Player.DashLockout = false
		c.Core.Player.DashCDExpirationFrame = c.Core.F + 0.8*60
		evt = c.Core.Log.NewEvent("dash lockout evaluation started", glog.LogCooldownEvent, c.Index)
	}

	evt.Write("lockout", c.Core.Player.DashLockout).
		Write("expiry", c.Core.Player.DashCDExpirationFrame-c.Core.F).
		Write("expiry_frame", c.Core.Player.DashCDExpirationFrame)
}

func (c *Character) QueueDashStaminaConsumption(p map[string]int) {
	// 終了時にスタミナを消費
	c.Core.Tasks.Add(func() {
		req := c.Core.Player.AbilStamCost(c.Index, action.ActionDash, p)
		c.Core.Player.UseStam(req, action.ActionDash)
	}, c.DashLength()-1)
}

func (c *Character) DashLength() int {
	switch c.CharBody {
	case info.BodyBoy, info.BodyLoli:
		return 21
	case info.BodyMale:
		return 19
	case info.BodyLady:
		return 22
	default:
		return 20
	}
}

func (c *Character) DashToJumpLength() int {
	switch c.CharBody {
	case info.BodyGirl, info.BodyLoli:
		return 4
	case info.BodyBoy:
		return 2
	default:
		return 3
	}
}

func (c *Character) Jump(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(player.XianyunAirborneBuff) {
		c.Core.Player.SetAirborne(player.AirborneXianyun)
		// 両手剣/弓/法器は4/8、片手剣/長柄武器は5/9
		lowPlunge := 4
		highPlunge := 8
		switch c.Weapon.Class {
		case info.WeaponClassSword, info.WeaponClassSpear:
			lowPlunge = 5
			highPlunge = 9
		}

		animLength := 60 // 高空/低空落下攻撃のジャンプの上限値
		return action.Info{
			Frames: func(a action.Action) int {
				switch a {
				case action.ActionLowPlunge:
					return lowPlunge
				case action.ActionHighPlunge:
					return highPlunge
				default:
					return animLength // AirborneXianyun 中は落下攻撃以外のアクションができないため、後にアクションエラーになることが想定される
				}
			},
			AnimationLength: animLength,
			CanQueueAfter:   lowPlunge, // 最速キャンセル
			State:           action.JumpState,
		}, nil
	}
	f := c.JumpLength()
	return action.Info{
		Frames:          func(action.Action) int { return f },
		AnimationLength: f,
		CanQueueAfter:   f,
		State:           action.JumpState,
	}, nil
}

func (c *Character) JumpLength() int {
	if c.Core.Player.LastAction.Type == action.ActionDash {
		switch c.CharBody {
		case info.BodyGirl, info.BodyBoy:
			return 34
		default:
			return 37
		}
	}
	switch c.CharBody {
	case info.BodyBoy, info.BodyGirl:
		return 31
	case info.BodyMale:
		return 28
	case info.BodyLady:
		return 32
	case info.BodyLoli:
		return 29
	default:
		return 30
	}
}

func (c *Character) Walk(p map[string]int) (action.Info, error) {
	f, ok := p["f"]
	if !ok {
		f = 1
	}
	return action.Info{
		Frames:          func(action.Action) int { return f },
		AnimationLength: f,
		CanQueueAfter:   f,
		State:           action.WalkState,
	}, nil
}
