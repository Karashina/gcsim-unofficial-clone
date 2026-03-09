package shortcut

// NormalizeSetNameはセットのショートカット名を正式名に解決する。
// 名前がショートカットマップに見つからない場合はそのまま返す。
func NormalizeSetName(name string) string {
	if key, ok := SetNameToKey[name]; ok {
		return key.String()
	}
	return name
}

// NormalizeWeaponNameは武器のショートカット名を正式名に解決する。
// 名前がショートカットマップに見つからない場合はそのまま返す。
func NormalizeWeaponName(name string) string {
	if key, ok := WeaponNameToKey[name]; ok {
		return key.String()
	}
	return name
}

// NormalizeNameはセット名、次に武器名の順でショートカット名を正式名に解決する。見つからない場合は入力をそのまま返す。
func NormalizeName(name string) string {
	if key, ok := SetNameToKey[name]; ok {
		return key.String()
	}
	if key, ok := WeaponNameToKey[name]; ok {
		return key.String()
	}
	return name
}
