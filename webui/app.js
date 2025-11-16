// Initialize charts storage
let charts = {};

document.addEventListener('DOMContentLoaded', function() {
    console.log('[WebUI] Initializing...');
    // Get textarea element
    const textarea = document.getElementById('config-editor');
    
    // Set default config - simpler version for reliable execution
    const defaultConfig = `ineffa char lvl=90/90 cons=0 talent=9,9,9;
ineffa add weapon="deathmatch" refine=1 lvl=90/90;
ineffa add set="gt" count=4;
ineffa add stats hp=4780 atk=311 em=187 atk%=0.466 cd=0.622;
ineffa add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*2 er=0.0551*2 em=19.82*4 cr=0.0331*10 cd=0.0662*12;

xingqiu char lvl=90/90 cons=6 talent=9,9,9;
xingqiu add weapon="favsword" refine=3 lvl=90/90;
xingqiu add set="sms" count=4;
xingqiu add stats hp=4780 atk=311 er=0.518 cr=0.311 hydro%=0.466;
xingqiu add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*5 er=0.0551*2 em=19.82*2 cr=0.0331*10 cd=0.0662*11;

options swap_delay=12 iteration=50;
target lvl=100 resist=0.1 radius=2 pos=2.1,1.5 hp=999999999;
energy every interval=480,720 amount=1;

active ineffa;

ineffa skill, burst;
xingqiu skill, burst, attack:3;
ineffa attack:5;
xingqiu attack:3;
ineffa skill, attack:5;`;
    
    textarea.value = defaultConfig;
    console.log('[WebUI] Default config loaded');
    
    // Add keyboard shortcuts
    textarea.addEventListener('keydown', function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            runSimulation();
        }
        // Tab key support
        if (e.key === 'Tab') {
            e.preventDefault();
            const start = this.selectionStart;
            const end = this.selectionEnd;
            this.value = this.value.substring(0, start) + '  ' + this.value.substring(end);
            this.selectionStart = this.selectionEnd = start + 2;
        }
    });
});

function switchTab(tabId) {
    // Hide all tab contents
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // Remove active class from all tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // Show selected tab content
    document.getElementById(tabId).classList.add('active');
    
    // Add active class to clicked button
    event.target.classList.add('active');
}

function clearErrorHighlights() {
    // No need to clear highlights in plain textarea
}

async function runSimulation() {
    console.log('[WebUI] Starting simulation...');
    const textarea = document.getElementById('config-editor');
    const config = textarea.value;
    const errorMsg = document.getElementById('error-message');
    const loading = document.getElementById('loading');
    const resultsContainer = document.getElementById('results-container');
    const runButton = document.querySelector('.btn-run');
    
    console.log('[WebUI] Config length:', config.length);
    
    // Hide previous results and errors
    errorMsg.style.display = 'none';
    resultsContainer.style.display = 'none';
    loading.style.display = 'block';
    runButton.disabled = true;
    clearErrorHighlights();
    
    try {
        console.log('[WebUI] Sending request to /api/simulate');
        const response = await fetch('/api/simulate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ config })
        });
        
        console.log('[WebUI] Response status:', response.status);
        
        loading.style.display = 'none';
        runButton.disabled = false;
        
        if (!response.ok) {
            const error = await response.json();
            console.error('[WebUI] Error response:', error);
            handleError(error);
            return;
        }
        
        const result = await response.json();
        console.log('[WebUI] Simulation result:', result);
        displayResults(result);
        
    } catch (err) {
        console.error('[WebUI] Exception:', err);
        loading.style.display = 'none';
        runButton.disabled = false;
        errorMsg.textContent = 'エラー: ' + err.message;
        errorMsg.style.display = 'block';
    }
}

function handleError(error) {
    console.log('[WebUI] Handling error:', error);
    const errorMsg = document.getElementById('error-message');
    let message = error.message || error.error || 'シミュレーションに失敗しました';
    
    if (error.parse_errors && error.parse_errors.length > 0) {
        message = 'パースエラー:\n';
        error.parse_errors.forEach(pe => {
            if (pe.line) {
                message += `行 ${pe.line}: ${pe.message}\n`;
            } else {
                message += `${pe.message}\n`;
            }
        });
    }
    
    errorMsg.textContent = message;
    errorMsg.style.display = 'block';
}

function displayResults(result) {
    console.log('[WebUI] Displaying results...');
    const resultsContainer = document.getElementById('results-container');
    resultsContainer.style.display = 'block';
    
    // Display statistics
    displayStatistics(result);
    
    // Display character information
    displayCharacters(result);
    
    // Display target information
    displayTargetInfo(result);
    
    // Display charts
    displayCharts(result);
    
    console.log('[WebUI] Results displayed successfully');
    
    // Scroll to results
    resultsContainer.scrollIntoView({ behavior: 'smooth' });
}

function displayStatistics(result) {
    console.log('[WebUI] Displaying statistics...');
    const stats = result.statistics || {};
    
    // Extract main statistics
    const dps = stats.dps?.mean || 0;
    const eps = stats.eps?.mean || 0;
    const rps = stats.rps?.mean || 0;
    const hps = stats.hps?.mean || 0;
    const shp = stats.shp?.mean || 0;
    const duration = stats.duration?.mean || result.simulator_settings?.duration || 0;
    
    console.log('[WebUI] Stats:', { dps, eps, rps, hps, shp, duration });
    
    document.getElementById('stat-dps').textContent = formatNumber(dps);
    document.getElementById('stat-eps').textContent = formatNumber(eps);
    document.getElementById('stat-rps').textContent = formatNumber(rps);
    document.getElementById('stat-hps').textContent = formatNumber(hps);
    document.getElementById('stat-shp').textContent = formatNumber(shp);
    document.getElementById('stat-dur').textContent = formatNumber(duration);
}

function displayCharacters(result) {
    console.log('[WebUI] Displaying characters...');
    const container = document.getElementById('characters-list');
    container.innerHTML = '';
    
    if (!result.character_details || result.character_details.length === 0) {
        container.innerHTML = '<p>キャラクター情報がありません</p>';
        return;
    }
    
    result.character_details.forEach(char => {
        const charDiv = document.createElement('div');
        charDiv.className = 'char-card';
        
        const name = char.name || 'Unknown';
        const level = char.level || 1;
        const maxLevel = char.max_level || 90;
        const constellation = char.cons || 0;
        const weapon = char.weapon?.name || 'Unknown';
        const weaponLevel = char.weapon?.level || 1;
        const weaponMaxLevel = char.weapon?.max_level || 90;
        const weaponRefine = char.weapon?.refine || 1;
        const talents = char.talents || {};
        
        // Talents display
        let talentsText = '-';
        if (talents.attack || talents.skill || talents.burst) {
            talentsText = `${talents.attack || 1}/${talents.skill || 1}/${talents.burst || 1}`;
        }
        
        // Sets display
        let setsHTML = '';
        if (char.sets && Object.keys(char.sets).length > 0) {
            const setsList = Object.entries(char.sets).map(([set, count]) => 
                `<span class="chip">${set} (${count})</span>`
            ).join(' ');
            setsHTML = `<div style="margin: 6px 0;"><strong>聖遺物セット:</strong> ${setsList}</div>`;
        }
        
        // Stats display with proper names
        let statsHTML = '';
        if (char.snapshot && char.snapshot.length > 0) {
            statsHTML = '<div style="margin-top: 8px;"><strong>ステータス詳細:</strong>';
            statsHTML += '<div style="display: grid; grid-template-columns: repeat(2, 1fr); gap: 4px; font-size: 0.85rem; margin-top: 4px;">';
            
            const statMapping = [
                { idx: 0, name: 'HP', format: (v) => Math.round(v) },
                { idx: 2, name: '基礎HP', format: (v) => Math.round(v) },
                { idx: 3, name: '攻撃力', format: (v) => Math.round(v) },
                { idx: 5, name: '基礎攻撃力', format: (v) => Math.round(v) },
                { idx: 4, name: '防御力', format: (v) => Math.round(v) },
                { idx: 7, name: '元素熟知', format: (v) => Math.round(v) },
                { idx: 9, name: '会心率', format: (v) => (v * 100).toFixed(1) + '%' },
                { idx: 10, name: '会心ダメージ', format: (v) => (v * 100).toFixed(1) + '%' },
                { idx: 8, name: '元素チャージ効率', format: (v) => (v * 100).toFixed(1) + '%' },
            ];
            
            statMapping.forEach(({idx, name, format}) => {
                if (char.snapshot[idx] !== undefined && char.snapshot[idx] !== 0) {
                    statsHTML += `<div class="info-row" style="padding: 2px 0;">
                        <span class="info-label">${name}:</span>
                        <span class="info-value">${format(char.snapshot[idx])}</span>
                    </div>`;
                }
            });
            statsHTML += '</div></div>';
        }
        
        charDiv.innerHTML = `
            <div class="char-name">${name}</div>
            <div class="char-info-compact">
                <div class="info-row">
                    <span class="info-label">Lv.</span>
                    <span class="info-value">${level}/${maxLevel}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">天賦Lv.</span>
                    <span class="info-value">${talentsText}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">星座</span>
                    <span class="info-value">C${constellation}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">武器</span>
                    <span class="info-value">${weapon} Lv.${weaponLevel}/${weaponMaxLevel} (R${weaponRefine})</span>
                </div>
            </div>
            ${setsHTML}
            ${statsHTML}
        `;
        
        container.appendChild(charDiv);
    });
}

function displayTargetInfo(result) {
    const container = document.getElementById('target-details');
    container.innerHTML = '';
    
    if (!result.target_details || result.target_details.length === 0) {
        container.innerHTML = '<p>ターゲット情報がありません</p>';
        return;
    }
    
    result.target_details.forEach((target, idx) => {
        const targetDiv = document.createElement('div');
        targetDiv.className = 'char-card';
        
        const name = target.name || `ターゲット ${idx + 1}`;
        const level = target.level || 1;
        const hp = target.hp || 0;
        
        let resistHTML = '';
        if (target.resist && Object.keys(target.resist).length > 0) {
            resistHTML = '<div style="margin-top: 10px;"><strong>元素耐性:</strong><br>';
            for (const [element, resist] of Object.entries(target.resist)) {
                resistHTML += `<div class="info-row">
                    <span class="info-label">${element}</span>
                    <span class="info-value">${(resist * 100).toFixed(1)}%</span>
                </div>`;
            }
            resistHTML += '</div>';
        }
        
        targetDiv.innerHTML = `
            <div class="char-name">${name}</div>
            <div class="info-row">
                <span class="info-label">レベル</span>
                <span class="info-value">${level}</span>
            </div>
            <div class="info-row">
                <span class="info-label">HP</span>
                <span class="info-value">${formatNumber(hp)}</span>
            </div>
            ${resistHTML}
        `;
        
        container.appendChild(targetDiv);
    });
}

function displayCharts(result) {
    console.log('[WebUI] Displaying charts...');
    // Destroy existing charts
    Object.values(charts).forEach(chart => {
        if (chart) chart.destroy();
    });
    charts = {};
    
    const stats = result.statistics || {};
    
    // Character DPS Chart (100% Stacked Bar Chart)
    if (stats.character_dps || (result.character_details && result.character_details.length > 0)) {
        const ctx = document.getElementById('char-dps-chart');
        const charDpsData = [];
        const charNames = [];
        
        if (result.character_details) {
            result.character_details.forEach((char, idx) => {
                charNames.push(char.name || `キャラ${idx+1}`);
                const dpsValue = stats.character_dps?.[idx]?.mean || 
                               stats.character_dps?.[char.name]?.mean || 0;
                charDpsData.push(dpsValue);
            });
        }
        
        if (charDpsData.length > 0) {
            charts.charDps = createStackedBarChart(ctx, ['チーム'], [charNames, charDpsData], 'キャラクター別DPS');
        }
    }
    
    // Source DPS Chart
    if (stats.dps_by_element || stats.source_dps) {
        const ctx = document.getElementById('source-dps-chart');
        let sourceData = {};
        
        if (stats.dps_by_element && stats.dps_by_element.length > 0) {
            // Extract from character DPS by element
            stats.dps_by_element.forEach((charData, idx) => {
                const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
                if (charData.elements) {
                    Object.entries(charData.elements).forEach(([element, data]) => {
                        const key = `${charName} (${element})`;
                        sourceData[key] = data.mean || data;
                    });
                }
            });
        }
        
        const data = extractChartData(sourceData);
        if (data.labels.length > 0) {
            charts.sourceDps = createBarChart(ctx, data.labels, data.values, 'ソース別DPS');
        }
    }
    
    // Damage Distribution Chart (Time-based line chart)
    if (stats.damage_buckets) {
        const ctx = document.getElementById('damage-dist-chart');
        const buckets = stats.damage_buckets;
        const bucketSize = buckets.bucket_size || 30;
        const bucketData = buckets.buckets || [];
        
        const timeLabels = bucketData.map((_, idx) => `${(idx * bucketSize).toFixed(0)}s`);
        const damageValues = bucketData.map(bucket => bucket.mean || 0);
        
        if (timeLabels.length > 0) {
            charts.damageDist = createLineChart(ctx, timeLabels, damageValues, 'ダメージ');
        }
    }
    
    // Energy Chart (Source-based)
    if (stats.total_source_energy && stats.total_source_energy.length > 0) {
        const ctx = document.getElementById('energy-chart');
        const energyData = {};
        
        stats.total_source_energy.forEach((charEnergy, idx) => {
            const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
            if (charEnergy && typeof charEnergy === 'object') {
                Object.entries(charEnergy).forEach(([source, value]) => {
                    energyData[`${charName}: ${source}`] = value;
                });
            }
        });
        
        const data = extractChartData(energyData);
        if (data.labels.length > 0) {
            charts.energy = createBarChart(ctx, data.labels, data.values, 'エネルギー');
        }
    }
    
    // Reaction Count Chart
    if (stats.source_reactions && stats.source_reactions.length > 0) {
        const ctx = document.getElementById('reaction-count-chart');
        const reactionData = {};
        
        stats.source_reactions.forEach((charReactions, idx) => {
            const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
            if (charReactions && typeof charReactions === 'object') {
                Object.entries(charReactions).forEach(([reaction, value]) => {
                    if (value && value !== 0) {
                        reactionData[`${charName}: ${reaction}`] = value;
                    }
                });
            }
        });
        
        const data = extractChartData(reactionData);
        if (data.labels.length > 0) {
            charts.reactions = createBarChart(ctx, data.labels, data.values, '反応回数');
        }
    }
    
    // Aura Uptime Chart
    if (stats.target_aura_uptime && stats.target_aura_uptime.length > 0) {
        const ctx = document.getElementById('aura-uptime-chart');
        const auraData = {};
        
        stats.target_aura_uptime.forEach((targetAura, idx) => {
            if (targetAura && typeof targetAura === 'object') {
                Object.entries(targetAura).forEach(([element, value]) => {
                    if (value && value !== 0) {
                        auraData[`ターゲット${idx+1}: ${element}`] = value * 100; // Convert to percentage
                    }
                });
            }
        });
        
        const data = extractChartData(auraData);
        if (data.labels.length > 0) {
            charts.aura = createBarChart(ctx, data.labels, data.values, '付着時間 (%)');
        }
    }
    
    console.log('[WebUI] Charts displayed');
}

function extractChartData(dataObj) {
    const labels = [];
    const values = [];
    
    if (typeof dataObj === 'object' && dataObj !== null) {
        for (const [key, value] of Object.entries(dataObj)) {
            if (typeof value === 'number') {
                labels.push(key);
                values.push(value);
            } else if (typeof value === 'object' && value.mean !== undefined) {
                labels.push(key);
                values.push(value.mean);
            }
        }
    }
    
    return { labels, values };
}

function createStackedBarChart(ctx, categories, [charNames, charValues], title) {
    // Calculate percentages
    const total = charValues.reduce((a, b) => a + b, 0);
    const percentages = charValues.map(v => total > 0 ? (v / total) * 100 : 0);
    
    const colors = [
        'rgba(102, 126, 234, 0.8)',
        'rgba(118, 75, 162, 0.8)',
        'rgba(237, 100, 166, 0.8)',
        'rgba(255, 154, 158, 0.8)',
    ];
    
    const datasets = charNames.map((name, idx) => ({
        label: name,
        data: [percentages[idx]],
        backgroundColor: colors[idx % colors.length],
        borderColor: colors[idx % colors.length].replace('0.8', '1'),
        borderWidth: 1
    }));
    
    return new Chart(ctx, {
        type: 'bar',
        data: {
            labels: categories,
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'bottom'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const charIdx = context.datasetIndex;
                            const dps = charValues[charIdx];
                            const pct = percentages[charIdx];
                            return `${context.dataset.label}: ${pct.toFixed(1)}% (DPS: ${Math.round(dps).toLocaleString('ja-JP')})`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    stacked: true
                },
                y: {
                    stacked: true,
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: function(value) {
                            return value + '%';
                        }
                    }
                }
            }
        }
    });
}

function createBarChart(ctx, labels, data, label) {
    return new Chart(ctx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [{
                label: label,
                data: data,
                backgroundColor: 'rgba(102, 126, 234, 0.6)',
                borderColor: 'rgba(102, 126, 234, 1)',
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

function createLineChart(ctx, labels, data, label) {
    return new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: label,
                data: data,
                backgroundColor: 'rgba(102, 126, 234, 0.2)',
                borderColor: 'rgba(102, 126, 234, 1)',
                borderWidth: 2,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: false
                }
            },
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

function formatNumber(num) {
    if (num === undefined || num === null) return '-';
    if (typeof num !== 'number') return num;
    
    // Format without K/M suffixes
    return Math.round(num).toLocaleString('ja-JP');
}

function formatStatName(statKey) {
    const statNames = {
        'hp': 'HP',
        'hp%': 'HP%',
        'atk': '攻撃力',
        'atk%': '攻撃力%',
        'def': '防御力',
        'def%': '防御力%',
        'em': '元素熟知',
        'er': '元素チャージ効率',
        'cr': '会心率',
        'cd': '会心ダメージ',
        'heal': '治療効果',
        'pyro%': '炎元素ダメージ',
        'hydro%': '水元素ダメージ',
        'cryo%': '氷元素ダメージ',
        'electro%': '雷元素ダメージ',
        'anemo%': '風元素ダメージ',
        'geo%': '岩元素ダメージ',
        'dendro%': '草元素ダメージ',
        'phys%': '物理ダメージ'
    };
    
    return statNames[statKey] || statKey;
}

function formatStatValue(statKey, value) {
    if (value === undefined || value === null) return '-';
    
    const percentStats = ['hp%', 'atk%', 'def%', 'er', 'cr', 'cd', 'heal', 
                          'pyro%', 'hydro%', 'cryo%', 'electro%', 'anemo%', 'geo%', 'dendro%', 'phys%'];
    
    if (percentStats.includes(statKey)) {
        return (value * 100).toFixed(1) + '%';
    } else {
        return value.toFixed(0);
    }
}
