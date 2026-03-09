package glog

type Source int

const (
	LogInvalid     Source = iota
	LogDamageEvent        // ダメージを追跡
	LogPreDamageMod
	LogCalc         // 詳細な計算
	LogElementEvent // 元素付与を追跡
	LogSnapshotEvent
	LogStatusEvent
	LogActionEvent
	LogEnergyEvent
	LogCharacterEvent
	LogEnemyEvent
	LogSimEvent
	LogArtifactEvent
	LogWeaponEvent
	LogShieldEvent
	LogConstructEvent
	LogICDEvent
	LogDebugEvent    // デバッグ用
	LogWarnings      // 問題が発生した場合
	LogPlayerEvent   // プレイヤー関連イベント（スタミナ、スワップCD、回復、被ダメージ等）
	LogHealEvent     // 回復イベント
	LogHurtEvent     // 被ダメージイベント
	LogCooldownEvent // クールダウンの開始/終了を追跡
	LogHitlagEvent   // ヒットラグのデバッグ用
	LogUserEvent     // ユーザー print イベント
)

var LogSourceFromString = map[string]Source{
	"":                LogInvalid,
	"damage":          LogDamageEvent,
	"pre_damage_mods": LogPreDamageMod,
	"calc":            LogCalc,
	"element":         LogElementEvent,
	"snapshot":        LogSnapshotEvent,
	"status":          LogStatusEvent,
	"action":          LogActionEvent,
	"energy":          LogEnergyEvent,
	"character":       LogCharacterEvent,
	"enemy":           LogEnemyEvent,
	"sim":             LogSimEvent,
	"artifact":        LogArtifactEvent,
	"weapon":          LogWeaponEvent,
	"shield":          LogShieldEvent,
	"construct":       LogConstructEvent,
	"icd":             LogICDEvent,
	"debug":           LogDebugEvent,
	"warning":         LogWarnings,
	"player":          LogPlayerEvent,
	"heal":            LogHealEvent,
	"hurt":            LogHurtEvent,
	"cooldown":        LogCooldownEvent,
	"hitlag":          LogHitlagEvent,
	"user":            LogUserEvent,
}

var LogSourceString = [...]string{
	"",
	"damage",
	"pre_damage_mods",
	"calc",
	"element",
	"snapshot",
	"status",
	"action",
	"energy",
	"character",
	"enemy",
	"sim",
	"artifact",
	"weapon",
	"shield",
	"construct",
	"icd",
	"debug",
	"warning",
	"player",
	"heal",
	"hurt",
	"cooldown",
	"hitlag",
	"user",
}

func (l Source) String() string {
	return LogSourceString[l]
}
