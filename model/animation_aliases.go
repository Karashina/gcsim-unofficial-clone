package model

import "github.com/genshinsim/gcsim/pkg/core/info"

// Compatibility aliases: map model animation delay keys to the canonical info ones.
// This is a small, temporary shim to let us delete the larger placeholder file
// while keeping existing callers building. We'll migrate callers to info.* later.

type AnimationDelayKey = info.AnimationDelayKey

const (
	AnimationXingqiuN0StartDelay = info.AnimationXingqiuN0StartDelay
	AnimationYelanN0StartDelay   = info.AnimationYelanN0StartDelay
)
