package shortcut

// NormalizeSetName resolves a set shortcut name to its canonical form.
// If the name is not found in the shortcut map, it is returned as-is.
func NormalizeSetName(name string) string {
	if key, ok := SetNameToKey[name]; ok {
		return key.String()
	}
	return name
}

// NormalizeWeaponName resolves a weapon shortcut name to its canonical form.
// If the name is not found in the shortcut map, it is returned as-is.
func NormalizeWeaponName(name string) string {
	if key, ok := WeaponNameToKey[name]; ok {
		return key.String()
	}
	return name
}

// NormalizeName resolves a shortcut name to its canonical form by trying
// set names first, then weapon names. Returns the input as-is if not found.
func NormalizeName(name string) string {
	if key, ok := SetNameToKey[name]; ok {
		return key.String()
	}
	if key, ok := WeaponNameToKey[name]; ok {
		return key.String()
	}
	return name
}
