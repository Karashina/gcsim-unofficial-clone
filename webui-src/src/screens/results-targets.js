/**
 * Results Screen - Target Information Module
 * Displays target/enemy information
 */

/**
 * Display target information
 * @param {Object} result - Simulation result object
 */
export function displayTargets(result) {
    console.log('[WebUI] Displaying targets...');
    const container = document.getElementById('target-info');
    if (!container) {
        console.warn('[WebUI] target-info container not found');
        return;
    }
    
    container.innerHTML = '';
    
    if (!result.target_details || result.target_details.length === 0) {
        container.innerHTML = '<p>ターゲット情報がありません</p>';
        return;
    }
    
    result.target_details.forEach((target, idx) => {
        const targetCard = createTargetCard(target, idx);
        container.appendChild(targetCard);
    });
}

/**
 * Create a single target card element
 * @param {Object} target - Target data
 * @param {number} idx - Target index
 * @returns {HTMLElement} Target card element
 */
function createTargetCard(target, idx) {
    const targetDiv = document.createElement('div');
    targetDiv.className = 'target-card';
    
    // Remove ~~strikethrough~~ tokens from strings
    const name = _stripStrikeTokens(target.name) || `ターゲット ${idx + 1}`;
    const level = target.level || 1;
    const hp = target.hp || 0;
    
    // Build resist lines
    let resistLines = '';
    if (target.resist && Object.keys(target.resist).length > 0) {
        for (const [element, resist] of Object.entries(target.resist)) {
            const el = _stripStrikeTokens(element);
            if (!el) continue; // Skip if entirely struck out
            resistLines += `
                <div class="info-row">
                    <span class="info-label">${el}</span>
                    <span class="info-value">${(resist * 100).toFixed(1)}%</span>
                </div>
            `;
        }
    }
    
    targetDiv.innerHTML = `
        <div class="target-header">
            <div class="target-name">${name}</div>
            <div class="target-level">Lv.${level}</div>
        </div>
        <div class="info-row">
            <span class="info-label">HP</span>
            <span class="info-value">${formatNumber(hp)}</span>
        </div>
        ${resistLines ? `
            <div style="margin-top: 8px; font-weight: 600; font-size: 0.85rem;">耐性:</div>
            ${resistLines}
        ` : ''}
    `;
    
    return targetDiv;
}

/**
 * Helper function to remove ~~strikethrough~~ tokens
 * Uses underscore prefix to avoid TDZ issues with external stripStrikeTokens
 * @param {string} s - String to process
 * @returns {string} Cleaned string
 */
function _stripStrikeTokens(s) {
    if (!s) return s;
    return s.replace(/~~.*?~~/g, '').trim();
}

/**
 * Format large numbers with thousands separators
 * @param {number} num - Number to format
 * @returns {string} Formatted number
 */
function formatNumber(num) {
    return Number(num).toLocaleString('ja-JP');
}
