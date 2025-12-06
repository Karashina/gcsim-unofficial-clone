/**
 * Localization Helper
 * Provides access to Japanese mappings for characters, weapons, and artifacts
 * These are loaded from jp_mappings.generated.js
 */

/**
 * Convert character key to Japanese name
 * @param {string} key - Character key
 * @returns {string} Japanese name or original key
 */
export function toJPCharacter(key) {
    if (!key) return key;
    return window.CHAR_TO_JP?.[key] || key;
}

/**
 * Convert weapon key to Japanese name
 * @param {string} key - Weapon key
 * @returns {string} Japanese name or original key
 */
export function toJPWeapon(key) {
    if (!key) return key;
    return window.WEAPON_TO_JP?.[key] || key;
}

/**
 * Convert artifact key to Japanese name
 * @param {string} key - Artifact set key
 * @returns {string} Japanese name or original key
 */
export function toJPArtifact(key) {
    if (!key) return key;
    return window.ARTIFACT_TO_JP?.[key] || key;
}
