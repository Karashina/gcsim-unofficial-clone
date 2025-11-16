package simulation

import (
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core"
	"github.com/Karashina/gcsim-unofficial-clone/pkg/core/info"
)

func NewCore(seed int64, debug bool, cfg *info.ActionList) (*core.Core, error) {
	return core.New(core.Opt{
		Seed:              seed,
		Debug:             debug,
		Delays:            cfg.Settings.Delays,
		DefHalt:           cfg.Settings.DefHalt,
		DamageMode:        cfg.Settings.DamageMode,
		EnableHitlag:      cfg.Settings.EnableHitlag,
		IgnoreBurstEnergy: cfg.Settings.IgnoreBurstEnergy,
	})
}

