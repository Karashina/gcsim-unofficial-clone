package conditional

import (
	"fmt"

	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/action"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/attributes"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/keys"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/player/character"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/shortcut"
)

func evalCharacter(c *core.Core, key keys.Char, fields []string) (any, error) {
	if err := fieldsCheck(fields, 2, "character"); err != nil {
		return 0, err
	}
	char, ok := c.Player.ByKey(key)
	if !ok {
		return 0, fmt.Errorf("character %v not in team when evaluating condition", key)
	}

	// special case for ability conditions. since fields are swapped
	// .kokomi.<abil>.<cond>
	typ := fields[1]
	act := action.StringToAction(typ)
	if act != action.InvalidAction {
		if err := fieldsCheck(fields, 3, "character ability"); err != nil {
			return 0, err
		}
		return evalCharacterAbil(c, char, act, fields[2])
	}

	charCat := "character " + typ

	switch typ {
	case "id":
		return int(char.Base.Key), nil
	case "cons":
		return char.Base.Cons, nil
	case "energy":
		return char.Energy, nil
	case "energymax":
		return char.EnergyMax, nil
	case "hp":
		return char.CurrentHP(), nil
	case "hpmax":
		return char.MaxHP(), nil
	case "hpratio":
		return char.CurrentHPRatio(), nil
	case "normal":
		return char.NextNormalCounter(), nil
	case "onfield":
		return c.Player.Active() == char.Index, nil
	case "weapon":
		// Return the canonical weapon name as a string for string comparison
		// e.g. .varka.weapon == "gestofthemightywolf"
		return char.Weapon.Key.String(), nil
	case "weaponkey":
		// Return the numeric weapon key for backward-compatible numeric comparison
		// e.g. .varka.weaponkey == .keys.weapon.gestofthemightywolf
		return int(char.Weapon.Key), nil
	case "set":
		// Return the canonical set name of the equipped set with 4+ pieces
		// e.g. .varka.set == "adcfrw"
		return evalCharacterMainSet(char), nil
	case "status":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return char.StatusDuration(fields[2]), nil
	case "mods":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return char.StatusDuration(fields[2]), nil
	case "infusion":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return c.Player.WeaponInfuseIsActive(char.Index, fields[2]), nil
	case "tags":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return char.Tag(fields[2]), nil
	case "stats":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return evalCharacterStats(char, fields[2])
	case "bol":
		return char.CurrentHPDebt(), nil
	case "bolratio":
		return char.CurrentHPDebtRatio(), nil
	case "sets":
		if err := fieldsCheck(fields, 3, charCat); err != nil {
			return 0, err
		}
		return evalCharacterSets(char, fields[2])
	default: // .kokomi.*
		return char.Condition(fields[1:])
	}
}

func evalCharacterStats(char *character.CharWrapper, stat string) (float64, error) {
	key := attributes.StrToStatType(stat)
	if key == -1 {
		return 0, fmt.Errorf("invalid stat key %v in character stat condition", stat)
	}
	return char.Stat(key), nil
}

func evalCharacterSets(char *character.CharWrapper, set string) (float64, error) {
	setKey, ok := shortcut.SetNameToKey[set]
	if !ok {
		return 0, fmt.Errorf("invalid set key %v in character set condition", set)
	}
	setInfo, ok := char.Equip.Sets[setKey]
	if !ok {
		return 0, nil
	}
	return float64(setInfo.GetCount()), nil
}

// evalCharacterMainSet returns the canonical name of the equipped set with the
// highest piece count (must be >= 4). Returns "" if no 4-piece set is equipped.
func evalCharacterMainSet(char *character.CharWrapper) string {
	var bestKey keys.Set
	bestCount := 0
	for k, v := range char.Equip.Sets {
		if c := v.GetCount(); c >= 4 && c > bestCount {
			bestKey = k
			bestCount = c
		}
	}
	if bestCount == 0 {
		return ""
	}
	return bestKey.String()
}

func evalCharacterAbil(c *core.Core, char *character.CharWrapper, act action.Action, typ string) (any, error) {
	switch typ {
	case "cd":
		if act == action.ActionSwap {
			return c.Player.SwapCD, nil
		}
		if act == action.ActionDash {
			if c.Player.Active() == char.Index && c.Player.DashLockout {
				return c.Player.DashCDExpirationFrame - c.F, nil
			}
			if c.Player.Active() != char.Index && char.DashLockout {
				return char.RemainingDashCD, nil
			}
			return 0, nil
		}
		return char.Cooldown(act), nil
	case "charge":
		return char.Charges(act), nil
	case "ready":
		if act == action.ActionSwap {
			return c.Player.SwapCD == 0 || c.Player.Active() == char.Index, nil
		}
		// TODO: nil map may cause problems here??
		ok, _ := char.ActionReady(act, nil)
		return ok, nil
	default:
		return 0, fmt.Errorf("bad character ability condition: invalid type %v", typ)
	}
}
