package attacks

// Temporary local aliases for ICD groups added upstream but missing in this fork.
// These are safe, minimal shims to avoid compile errors until generated files are synced.
// They intentionally map to existing groups (Default) to preserve behavior until proper ICDGroup
// entries are imported.

const (
    // Upstream added this specialized group; map to Default for now.
    ICDGroupAinoBurstEnhanced ICDGroup = ICDGroupDefault
)
