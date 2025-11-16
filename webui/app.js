// Initialize editor
let editor;
let charts = {};

document.addEventListener('DOMContentLoaded', function() {
    // Get textarea element
    const textarea = document.getElementById('config-editor');
    
    // Set default config
    const defaultConfig = `ineffa char lvl=90/90 cons=0 talent=9,9,9;
ineffa add weapon="deathmatch" refine=1 lvl=90/90;
ineffa add set="gt" count=4;
ineffa add stats hp=4780 atk=311 em=187 atk%=0.466 cd=0.622;
ineffa add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*2 er=0.0551*2 em=19.82*4 cr=0.0331*10 cd=0.0662*12;

sucrose char lvl=90/90 cons=6 talent=9,9,9;
sucrose add weapon="hakushin" refine=5 lvl=90/90;
sucrose add set="vv" count=5;
sucrose add stats hp=4780 atk=311 em=560;
sucrose add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*3 er=0.0551*3 em=19.82*6 cr=0.0331*11 cd=0.0662*7;

flins char lvl=90/90 cons=0 talent=9,9,9;
flins add weapon="bloodsoakedruins" refine=1 lvl=90/90;
flins add set="notsu" count=4;
flins add stats hp=4780 atk=311 atk%=0.466 atk%=0.466 cd=0.622;
flins add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*5 er=0.0551*2 em=19.82*2 cr=0.0331*9 cd=0.0662*12;

xingqiu char lvl=90/90 cons=6 talent=9,9,9;
xingqiu add weapon="favsword" refine=3 lvl=90/90;
xingqiu add set="sms" count=4;
xingqiu add stats hp=4780 atk=311 er=0.518 cr=0.311 hydro%=0.466;
xingqiu add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*5 er=0.0551*2 em=19.82*2 cr=0.0331*10 cd=0.0662*11;

options swap_delay=12 iteration=1000;
target lvl=100 resist=0.1 radius=2 pos=2.1,1.5 hp=999999999;
energy every interval=480,720 amount=1;

active ineffa;

for let i=0; i<4; i=i+1 {
  let r = rand();
  ineffa skill;
  if .ineffa.burst.ready && .ineffa.energymax {
    ineffa burst;
  }
  xingqiu skill, dash, burst, attack:1;
  sucrose attack, skill, jump;
  flins skill, skill, burst;
  while !.flins.northlandup {
    flins attack;
    if .flins.normal < 4 {
      flins dash;
    }
  }
  flins skill, burst;
  while !.flins.northlandup {
    flins attack;
    if .flins.normal < 4 {
      flins dash;
    }
  }
  flins skill, burst;
  flins attack:4, dash, attack:5;
}`;
    
    textarea.value = defaultConfig;
    
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
    const textarea = document.getElementById('config-editor');
    const config = textarea.value;
    const errorMsg = document.getElementById('error-message');
    const loading = document.getElementById('loading');
    const resultsContainer = document.getElementById('results-container');
    const runButton = document.querySelector('.btn-run');
    
    // Hide previous results and errors
    errorMsg.style.display = 'none';
    resultsContainer.style.display = 'none';
    loading.style.display = 'block';
    runButton.disabled = true;
    clearErrorHighlights();
    
    try {
        const response = await fetch('/api/simulate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ config })
        });
        
        loading.style.display = 'none';
        runButton.disabled = false;
        
        if (!response.ok) {
            const error = await response.json();
            handleError(error);
            return;
        }
        
        const result = await response.json();
        displayResults(result);
        
    } catch (err) {
        loading.style.display = 'none';
        runButton.disabled = false;
        errorMsg.textContent = 'エラー: ' + err.message;
        errorMsg.style.display = 'block';
    }
}

function handleError(error) {
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
    
    // Scroll to results
    resultsContainer.scrollIntoView({ behavior: 'smooth' });
}

function displayStatistics(result) {
    const stats = result.statistics || {};
    
    // Extract main statistics
    const dps = stats.dps?.mean || 0;
    const eps = stats.eps?.mean || 0;
    const rps = stats.rps?.mean || 0;
    const hps = stats.hps?.mean || 0;
    const shp = stats.shp?.mean || 0;
    const duration = stats.duration?.mean || result.simulator_settings?.duration || 0;
    
    document.getElementById('stat-dps').textContent = formatNumber(dps);
    document.getElementById('stat-eps').textContent = formatNumber(eps);
    document.getElementById('stat-rps').textContent = formatNumber(rps);
    document.getElementById('stat-hps').textContent = formatNumber(hps);
    document.getElementById('stat-shp').textContent = formatNumber(shp);
    document.getElementById('stat-dur').textContent = formatNumber(duration);
}

function displayCharacters(result) {
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
        const weaponRefine = char.weapon?.refine || 1;
        
        let statsHTML = '<div style="margin-top: 10px;">';
        statsHTML += '<strong>ステータス:</strong><br>';
        
        if (char.stats) {
            for (const [key, value] of Object.entries(char.stats)) {
                if (value && value !== 0) {
                    statsHTML += `<div class="info-row">
                        <span class="info-label">${formatStatName(key)}</span>
                        <span class="info-value">${formatStatValue(key, value)}</span>
                    </div>`;
                }
            }
        }
        statsHTML += '</div>';
        
        let setsHTML = '';
        if (char.sets && Object.keys(char.sets).length > 0) {
            setsHTML = '<div style="margin-top: 10px;"><strong>聖遺物セット:</strong><br>';
            for (const [set, count] of Object.entries(char.sets)) {
                setsHTML += `<span class="chip">${set} (${count})</span> `;
            }
            setsHTML += '</div>';
        }
        
        charDiv.innerHTML = `
            <div class="char-name">${name}</div>
            <div class="info-row">
                <span class="info-label">レベル</span>
                <span class="info-value">${level}/${maxLevel}</span>
            </div>
            <div class="info-row">
                <span class="info-label">凸数</span>
                <span class="info-value">C${constellation}</span>
            </div>
            <div class="info-row">
                <span class="info-label">武器</span>
                <span class="info-value">${weapon} (R${weaponRefine})</span>
            </div>
            ${statsHTML}
            ${setsHTML}
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
    // Destroy existing charts
    Object.values(charts).forEach(chart => {
        if (chart) chart.destroy();
    });
    charts = {};
    
    const stats = result.statistics || {};
    
    // Character DPS Chart
    if (stats.character_dps) {
        const ctx = document.getElementById('char-dps-chart');
        const data = extractChartData(stats.character_dps);
        charts.charDps = createBarChart(ctx, data.labels, data.values, 'DPS');
    }
    
    // Source DPS Chart
    if (stats.source_dps || stats.damage_by_source) {
        const ctx = document.getElementById('source-dps-chart');
        const sourceData = stats.source_dps || stats.damage_by_source || {};
        const data = extractChartData(sourceData);
        charts.sourceDps = createBarChart(ctx, data.labels, data.values, 'DPS');
    }
    
    // Damage Distribution Chart
    if (stats.damage_buckets || stats.dps_by_target) {
        const ctx = document.getElementById('damage-dist-chart');
        const distData = stats.damage_buckets || stats.dps_by_target || {};
        const data = extractChartData(distData);
        charts.damageDist = createLineChart(ctx, data.labels, data.values, 'ダメージ');
    }
    
    // Energy Chart
    if (stats.energy_stats || stats.particle_count) {
        const ctx = document.getElementById('energy-chart');
        const energyData = stats.energy_stats || {};
        const data = extractChartData(energyData);
        charts.energy = createBarChart(ctx, data.labels, data.values, 'エネルギー');
    }
    
    // Reaction Count Chart
    if (stats.reactions || stats.reaction_count) {
        const ctx = document.getElementById('reaction-count-chart');
        const reactionData = stats.reactions || stats.reaction_count || {};
        const data = extractChartData(reactionData);
        charts.reactions = createBarChart(ctx, data.labels, data.values, '回数');
    }
    
    // Aura Uptime Chart
    if (stats.element_uptime || stats.aura_uptime) {
        const ctx = document.getElementById('aura-uptime-chart');
        const auraData = stats.element_uptime || stats.aura_uptime || {};
        const data = extractChartData(auraData);
        charts.aura = createBarChart(ctx, data.labels, data.values, '時間 (%)');
    }
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
    
    if (num >= 1000000) {
        return (num / 1000000).toFixed(2) + 'M';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(2) + 'K';
    } else {
        return num.toFixed(2);
    }
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
