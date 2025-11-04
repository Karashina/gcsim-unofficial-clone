package model

// Temporary shim for missing generated animation delay keys.
// These are minimal placeholders to let the repository compile during the staged
// refactor integration. They should be replaced with the real generated values
// from upstream main when available.

type AnimationDelayKey int

const (
	// Common animation delay keys observed in first-pass build logs
	AnimationXingqiuN0StartDelay AnimationDelayKey = iota
	AnimationYelanN0StartDelay
	// Generic placeholders for other characters; add more as needed
	AnimationPlaceholderA
	AnimationPlaceholderB
)
