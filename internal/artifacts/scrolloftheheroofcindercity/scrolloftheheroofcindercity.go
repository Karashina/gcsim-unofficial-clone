package scrolloftheheroofcindercity

import (
	"fmt"
	"slices"

	"github.com/Karashina/gcsim-unofficial-clone/internal/template/nightsoul"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/combat"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/event"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/reactions"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/targets"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/modifier"
)

var reactToElements = map[reactions.ReactionType][]attributes.Element{
	reactions.Overload:           {attributes.Electro, attributes.Pyro},
	reactions.Superconduct:       {attributes.Electro, attributes.Cryo},
	reactions.Melt:               {attributes.Pyro, attributes.Cryo},
	reactions.Vaporize:           {attributes.Pyro, attributes.Hydro},
	reactions.Freeze:             {attributes.Cryo, attributes.Hydro},
	reactions.ElectroCharged:     {attributes.Electro, attributes.Hydro},
	reactions.SwirlHydro:         {attributes.Anemo, attributes.Hydro},
	reactions.SwirlCryo:          {attributes.Anemo, attributes.Cryo},
	reactions.SwirlElectro:       {attributes.Anemo, attributes.Electro},
	reactions.SwirlPyro:          {attributes.Anemo, attributes.Pyro},
	reactions.CrystallizeHydro:   {attributes.Geo, attributes.Hydro},
	reactions.CrystallizeCryo:    {attributes.Geo, attributes.Cryo},
	reactions.CrystallizeElectro: {attributes.Geo, attributes.Electro},
	reactions.CrystallizePyro:    {attributes.Geo, attributes.Pyro},
	reactions.Aggravate:          {attributes.Dendro, attributes.Electro},
	reactions.Spread:             {attributes.Dendro},
	reactions.Quicken:            {attributes.Dendro, attributes.Electro},
	reactions.Bloom:              {attributes.Dendro, attributes.Hydro},
	reactions.Hyperbloom:         {attributes.Dendro, attributes.Electro},
	reactions.Burgeon:            {attributes.Dendro, attributes.Pyro},
	reactions.Burning:            {attributes.Dendro, attributes.Pyro},
}

func init() {
	core.RegisterSetFunc(keys.ScrollOfTheHeroOfCinderCity, NewSet)
}

type Set struct {
	Index int
	Count int

	c             *core.Core
	char          *character.CharWrapper
	buff          []float64
	nightsoulBuff []float64
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func (s *Set) buffCB(react reactions.ReactionType, gadgetEmit bool) func(args ...interface{}) bool {
	return func(args ...interface{}) bool {
		trg := args[0].(combat.Target)
		if gadgetEmit && trg.Type() != targets.TargettableGadget {
			return false
		}
		if !gadgetEmit && trg.Type() != targets.TargettableEnemy {
			return false
		}

		ae := args[1].(*combat.AttackEvent)
		if ae.Info.ActorIndex != s.char.Index {
			return false
		}

		hasNightsoul := s.char.StatusIsActive(nightsoul.NightsoulBlessingStatus)
		for _, other := range s.c.Player.Chars() {
			elements := reactToElements[react]
			for _, ele := range elements {
				stat := attributes.EleToDmgP(ele)
				other.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(fmt.Sprintf("scroll-4pc-%v", ele), 15*60),
					AffectedStat: stat,
					Amount: func() ([]float64, bool) {
						clear(s.buff)
						s.buff[stat] = 0.12
						return s.buff, true
					},
				})

				if !hasNightsoul {
					continue
				}
				other.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(fmt.Sprintf("scroll-4pc-nightsoul-%v", ele), 20*60),
					AffectedStat: stat,
					Amount: func() ([]float64, bool) {
						clear(s.nightsoulBuff)
						s.nightsoulBuff[stat] = 0.28
						return s.nightsoulBuff, true
					},
				})
			}
		}
		return false
	}
}

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		Count:         count,
		c:             c,
		char:          char,
		buff:          make([]float64, attributes.EndStatType),
		nightsoulBuff: make([]float64, attributes.EndStatType),
	}
	// 2セット: 近くのパーティメンバーがNightsoul Burstを発動すると、装備
	// キャラクターの元素エネルギーが6回復。
	if count >= 2 {
		c.Combat.Events.Subscribe(event.OnNightsoulBurst, func(args ...interface{}) bool {
			char.AddEnergy("scroll-2pc", 6)
			return false
		}, fmt.Sprintf("scroll-2pc-%v", char.Base.Key.String()))
	}
	// 4セット: 装備キャラが自身の元素タイプに関連する反応を発動すると、
	// 全周囲のパーティメンバーが反応に関与した元素の元素ダメージ+12%を15秒間獲得。
	// 装備キャラが夜魂の祝福状態の場合、
	// さらに元素ダメージ+28%を20秒間追加獲得。
	// 装備キャラはフィールド外でも発動可能。
	// 同名の聖遺物セットのダメージボーナスは重複しない。
	if count >= 4 {
		for evt, react := range map[event.Event]reactions.ReactionType{
			event.OnOverload:           reactions.Overload,
			event.OnSuperconduct:       reactions.Superconduct,
			event.OnMelt:               reactions.Melt,
			event.OnVaporize:           reactions.Vaporize,
			event.OnFrozen:             reactions.Freeze,
			event.OnElectroCharged:     reactions.ElectroCharged,
			event.OnSwirlHydro:         reactions.SwirlHydro,
			event.OnSwirlCryo:          reactions.SwirlCryo,
			event.OnSwirlElectro:       reactions.SwirlElectro,
			event.OnSwirlPyro:          reactions.SwirlPyro,
			event.OnCrystallizeHydro:   reactions.CrystallizeHydro,
			event.OnCrystallizeCryo:    reactions.CrystallizeCryo,
			event.OnCrystallizeElectro: reactions.CrystallizeElectro,
			event.OnCrystallizePyro:    reactions.CrystallizePyro,
			event.OnAggravate:          reactions.Aggravate,
			event.OnSpread:             reactions.Spread,
			event.OnQuicken:            reactions.Quicken,
			event.OnBloom:              reactions.Bloom,
			event.OnHyperbloom:         reactions.Hyperbloom,
			event.OnBurgeon:            reactions.Burgeon,
			event.OnBurning:            reactions.Burning,
		} {
			elements := reactToElements[react]
			if !slices.Contains(elements, char.Base.Element) {
				continue
			}
			gadgetEmit := false
			switch react {
			case reactions.Burgeon, reactions.Hyperbloom:
				gadgetEmit = true
			}
			c.Combat.Events.Subscribe(evt, s.buffCB(react, gadgetEmit), fmt.Sprintf("scroll-4pc-%v-%v", react, char.Base.Key.String()))
		}
	}

	return &s, nil
}
