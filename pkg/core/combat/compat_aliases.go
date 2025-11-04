package combat

import "github.com/genshinsim/gcsim/pkg/core/info"

// Compatibility aliases to map older combat.* identifiers to the new info.* types
// This file is a temporary shim to ease upstream refactor integration.

type AttackInfo = info.AttackInfo
type Snapshot = info.Snapshot
type AttackEvent = info.AttackEvent
type AttackCB = info.AttackCB
type AttackCBFunc = info.AttackCBFunc
type AttackPattern = info.AttackPattern
type Target = info.Target

// Provide a reasonable default for Reactable weapon gadget type used in some chars
// Upstream renamed gadget types; map the legacy name to a test gadget type to allow compilation.
var GadgetTypReactableweapon info.GadgetTyp = info.GadgetTypTest

// Additional temporary gadget aliases observed as missing during build
var GadgetTypBaronBunny info.GadgetTyp = info.GadgetTypTest
