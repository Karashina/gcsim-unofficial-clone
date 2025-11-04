package reactable

import "github.com/genshinsim/gcsim/pkg/core/info"

// Minimal shim for reactable package to provide commonly referenced enums
// and values used in character code during staged refactor.

type Element int

const (
	Electro Element = iota
	Pyro
	Cryo
	Hydro
	Anemo
	Dendro
	Geo
)

// Make Duration an alias to info.Durability so comparisons with info.Durability
// compile without mismatched-type errors.
type Duration = info.Durability

var ZeroDur Duration = 0
