/**
 * Results Screen - Character Display Module
 * Handles rendering of character cards with stats, weapons, artifacts
 */

import { toJPCharacter, toJPWeapon, toJPArtifact } from '../localization.js';

/**
 * Display character information cards
 * @param {Object} result - Simulation result object
 */
export function displayCharacters(result) {
    console.log('[WebUI] Displaying characters...');
    const container = document.getElementById('characters-list');
    if (!container) {
        console.error('[WebUI] characters-list container not found');
        return;
    }
    
    container.innerHTML = '';
    
    if (!result.character_details || result.character_details.length === 0) {
        container.innerHTML = '<p>キャラクター情報がありません</p>';
        return;
    }
    
    const gridDiv = document.createElement('div');
    gridDiv.className = 'characters-grid';
    
    result.character_details.forEach((char, idx) => {
        const charCard = createCharacterCard(char, idx);
        gridDiv.appendChild(charCard);
    });
    
    container.appendChild(gridDiv);
}

/**
 * Create a single character card element
 * @param {Object} char - Character data
 * @param {number} idx - Character index
 * @returns {HTMLElement} Character card element
 */
function createCharacterCard(char, idx) {
    console.log(`[WebUI] Character ${idx}:`, char.name);
    
    const charDiv = document.createElement('div');
    charDiv.className = 'char-card';
    
    // Extract character data
    const rawName = char.name || 'Unknown';
    const name = toJPCharacter(rawName);
    const level = char.level || 1;
    const maxLevel = char.max_level || 90;
    const constellation = char.cons || 0;
    
    // Talents
    const talents = char.talents || {};
    const talentsText = (talents.attack || talents.skill || talents.burst)
        ? `${talents.attack || 1}/${talents.skill || 1}/${talents.burst || 1}`
        : '-';
    
    // Weapon
    const weapon = char.weapon?.name || 'Unknown';
    const weaponJP = toJPWeapon(weapon);
    const weaponLevel = char.weapon?.level || 1;
    const weaponMaxLevel = char.weapon?.max_level || 90;
    const weaponRefine = char.weapon?.refine || 1;
    
    // Artifact set
    const artifactBadge = createArtifactBadge(char.sets);
    
    // Stats
    const statsHTML = createStatsHTML(char, name);
    
    // Build card HTML
    charDiv.innerHTML = `
        <div class="char-header">
            <div class="char-name-line">
                ${name}
                <span class="char-constellation">C${constellation}</span>
            </div>
            <div class="char-talents">${talentsText}</div>
        </div>
        <div class="char-subheader">
            <div class="char-en-name">${rawName}</div>
            <div class="char-level">Lv. ${level}/${maxLevel}</div>
        </div>
        ${artifactBadge}
        <div class="char-weapon">
            <div class="char-weapon-name">${weaponJP} Lv.${weaponLevel}/${weaponMaxLevel} (R${weaponRefine})</div>
            <div class="char-weapon-en">${weapon}</div>
        </div>
        <div class="char-stats">
            <div class="char-stats-title">ステータス詳細:</div>
            <div class="char-stats-list">
                ${statsHTML}
            </div>
        </div>
    `;
    
    return charDiv;
}

/**
 * Create artifact set badge HTML
 * @param {Object} sets - Character's artifact sets
 * @returns {string} Badge HTML
 */
function createArtifactBadge(sets) {
    if (!sets || Object.keys(sets).length === 0) {
        return '<div class="char-artifact"></div>';
    }
    
    const firstSet = Object.entries(sets)[0];
    const [setKey, count] = firstSet;
    const setName = toJPArtifact(setKey);
    
    return `
        <div class="char-artifact">
            <span class="chip">
                ${setName} (${count})
                <div class="small-en">${setKey}</div>
            </span>
        </div>
    `;
}

/**
 * Create stats HTML from snapshot_stats
 * @param {Object} char - Character data
 * @param {string} name - Character name (for logging)
 * @returns {string} Stats HTML
 */
function createStatsHTML(char, name) {
    const snapshotStats = char.snapshot_stats || char.snapshot || [];
    
    if (!snapshotStats || snapshotStats.length === 0) {
        console.log('[WebUI] No snapshot_stats for:', name);
        return '';
    }
    
    console.log(`[WebUI] ${name} snapshot_stats:`, snapshotStats);
    
    // Extract stats from snapshot_stats array
    const stats = {
        hp: snapshotStats[3] || 0,
        atk: snapshotStats[5] || 0,
        def: snapshotStats[2] || 0,
        em: snapshotStats[8] || 0,
        cr: snapshotStats[9] || 0,
        cd: snapshotStats[10] || 0,
        er: snapshotStats[7] || 0
    };
    
    const statDefs = [
        { name: 'HP', value: stats.hp, format: (v) => Math.round(v) },
        { name: '攻撃力', value: stats.atk, format: (v) => Math.round(v) },
        { name: '防御力', value: stats.def, format: (v) => Math.round(v) },
        { name: '元素熟知', value: stats.em, format: (v) => Math.round(v) },
        { name: '会心率', value: stats.cr, format: (v) => (v * 100).toFixed(1) + '%' },
        { name: '会心ダメージ', value: stats.cd, format: (v) => (v * 100).toFixed(1) + '%' },
        { name: '元素チャージ効率', value: stats.er, format: (v) => (v * 100).toFixed(1) + '%' }
    ];
    
    let html = '';
    statDefs.forEach(({ name: statName, value, format }) => {
        if (value !== undefined && value !== 0) {
            html += `
                <div class="info-row">
                    <span class="info-label">${statName}</span>
                    <span class="info-value">${format(value)}</span>
                </div>
            `;
        }
    });
    
    return html;
}
