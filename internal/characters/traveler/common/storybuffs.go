package common

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

func TravelerStoryBuffs(c *character.CharWrapper, p info.CharacterProfile) {
	// TravelerStoryBuffs はストーリークエストのクリアに基づいてバフを適用する
	//
	// base_atk_buff
	// 		第三章：第一幕（スメール魔神任務）クリアで獲得するバフ
	// 		+3 基礎攻撃力
	// skirk_story_buff
	// 		結晶の章：第一幕（スカーク伝説任務）クリアで獲得するバフ
	// 		+7 基礎攻撃力, +15 元素熟知, +50 基礎HP
	//
	// 全バフはデフォルトで有効
	baseAtkBuff, okBaseAtkBuff := p.Params["base_atk_buff"]
	skirkBuff, okSkirkBuff := p.Params["skirk_story_buff"]
	if !okBaseAtkBuff {
		baseAtkBuff = 1
	}
	if !okSkirkBuff {
		skirkBuff = 1 // デフォルトで最大バフ
	}

	m := make([]float64, attributes.EndStatType)
	if baseAtkBuff == 1 {
		m[attributes.BaseATK] += 3
	}
	if skirkBuff == 1 {
		m[attributes.BaseATK] += 7
		m[attributes.EM] += 15
		m[attributes.BaseHP] += 50
	}
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBase("traveler-story-quest-buffs", -1),
		Amount: func() ([]float64, bool) {
			return m, true
		},
	})
}
