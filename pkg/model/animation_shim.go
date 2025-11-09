package model

// Temporary shim for missing generated animation delay keys under pkg/model.
// These placeholders let the repo compile during staged integration. Replace
// with upstream-generated values from main when available.

type AnimationDelayKey int

const (
	AnimationXingqiuN0StartDelay AnimationDelayKey = iota
	AnimationYelanN0StartDelay
	AnimationPlaceholderA
	AnimationPlaceholderB
)
