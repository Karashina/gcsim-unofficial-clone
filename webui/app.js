// Initialize charts storage
let charts = {};
// Global reference to CodeMirror editor (if initialized)
var cmEditor = null;

// Global helper to convert hex color to rgba string with alpha
function hexToRgba(hex, alpha) {
    if (!hex) return `rgba(0,0,0,${alpha})`;
    const h = hex.replace('#', '');
    const bigint = parseInt(h, 16);
    const r = (bigint >> 16) & 255;
    const g = (bigint >> 8) & 255;
    const b = bigint & 255;
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

// Global character palette (kept consistent across charts)
const CHAR_PALETTE = ['#4e79a7','#59a14f','#f28e2b','#e15759','#76b7b2','#edc949','#af7aa1','#ff9da7'];
function getCharColor(i) {
    return CHAR_PALETTE[i % CHAR_PALETTE.length];
}

// Debug flag: set to true to enable verbose debug logging
const DEBUG = false;
function debugLog(...args) { if (DEBUG && console && console.log) console.log(...args); }

// Small deterministic string hash used to derive a hue for element coloring
function hashCode(str) {
    let h = 0;
    if (!str) return h;
    for (let i = 0; i < str.length; i++) {
        h = ((h << 5) - h) + str.charCodeAt(i);
        h |= 0; // convert to 32bit int
    }
    return h;
}

// Helper: extract numeric value from Chart.js tooltip context in a version-agnostic way
function ctxRawValue(context) {
    if (context === undefined || context === null) return 0;
    if (typeof context.raw === 'number') return context.raw;
    if (context.parsed) return context.parsed.x || context.parsed || 0;
    return context.raw || 0;
}

// Helper: set canvas drawing buffer and reserve parent height to avoid layout shifts
function setCanvasVisualSize(ctx, desiredHeightPx, minWidth = 300) {
    try {
        const canvas = ctx && ctx.canvas ? ctx.canvas : null;
        if (!canvas) return;
        const parent = canvas.parentElement;
        const dpr = (typeof window !== 'undefined' && window.devicePixelRatio) ? window.devicePixelRatio : 1;
        try { canvas.style.width = '100%'; } catch (e) {}
        const rect = parent && parent.getBoundingClientRect ? parent.getBoundingClientRect() : canvas.getBoundingClientRect();
        const visualWidth = rect && rect.width ? Math.max(minWidth, Math.floor(rect.width)) : Math.max(minWidth, Math.floor(canvas.offsetWidth || 600));
        const heightPx = Math.max(120, Math.floor(desiredHeightPx || 140));
        try { if (parent) parent.style.setProperty('min-height', heightPx + 'px', 'important'); } catch (e) {}
        canvas.width = Math.floor(visualWidth * dpr);
        canvas.height = Math.floor(heightPx * dpr);
        ensureContainerHeight(ctx, heightPx);
    } catch (e) { /* ignore sizing errors */ }
}

// Format numeric values for tooltips consistently. If value is integral, show integer; otherwise fixed decimals.
function formatValue(v, decimals = 2, suffix = '') {
    const n = Number(v) || 0;
    const isInt = Math.abs(n - Math.round(n)) < 1e-9;
    const body = isInt ? Math.round(n).toString() : n.toFixed(decimals);
    return body + (suffix || '');
}

// Extract a damage/DPS numeric value from various possible statistic shapes.
// Prefers .mean, then .damage, then .dps, then numeric leaves. Avoid returning
// pure "count/instances" objects by detecting key names like 'count' or 'instances'.
function extractDamageValue(v) {
    if (v === null || v === undefined) return 0;
    if (typeof v === 'number') return v;
    if (typeof v === 'object') {
        if (typeof v.mean === 'number') return v.mean;
        if (typeof v.damage === 'number') return v.damage;
        if (typeof v.dps === 'number') return v.dps;

        // Aggregate numeric leaves but avoid using pure count-only structures.
        let sum = 0;
        let numericCount = 0;
        let countLikeOnly = true;
        for (const [k, val] of Object.entries(v)) {
            if (typeof val === 'number') {
                sum += val;
                numericCount++;
                // if key doesn't look like a count, mark as not count-only
                if (!/count|counts|instances|uses|times|occur/i.test(k)) countLikeOnly = false;
            } else if (typeof val === 'object') {
                const nested = extractDamageValue(val);
                if (nested && typeof nested === 'number') {
                    sum += nested;
                    numericCount++;
                    countLikeOnly = false;
                }
            }
        }
        if (numericCount === 0) return 0;
        if (countLikeOnly) {
            // Looks like we only found counts; don't treat counts as damage
            return 0;
        }
        return sum;
    }
    return 0;
}

// Convert various statistic shapes into DPS (damage per second).
// If the incoming object already contains a dps field, use it directly.
// If it contains mean/damage as total damage per iteration, divide by durationMean.
// Return DescriptiveStats.mean when present. Do not attempt to convert totals to DPS by dividing
// by duration here — mean is the canonical numeric statistic returned by the simulator.
function extractDescriptiveMean(v) {
    if (v === null || v === undefined) return null;
    if (typeof v === 'object' && typeof v.mean === 'number') return v.mean;
    return null;
}

// Generic numeric extractor: prefer DescriptiveStats.mean, then number, then sum numeric leaves.
function extractNumber(v) {
    if (v === null || v === undefined) return 0;
    if (typeof v === 'number') return Number.isFinite(v) ? v : 0;
    if (typeof v === 'object') {
        if (typeof v.mean === 'number' && Number.isFinite(v.mean)) return v.mean;
        if (typeof v.value === 'number' && Number.isFinite(v.value)) return v.value;
        // sum numeric leaves
        let s = 0;
        let found = false;
        for (const vv of Object.values(v)) {
            if (typeof vv === 'number' && Number.isFinite(vv)) { s += vv; found = true; }
            else if (typeof vv === 'object') {
                const nested = extractNumber(vv);
                if (nested) { s += nested; found = true; }
            }
        }
        return found ? s : 0;
    }
    return 0;
}

// Ensure Chart.js (if loaded) uses UDEV Gothic as default font for canvas text
try {
    if (typeof Chart !== 'undefined' && Chart.defaults && Chart.defaults.font) {
        Chart.defaults.font.family = "'UDEV Gothic', 'Segoe UI', Arial, sans-serif";
    }
} catch (e) {
    // no-op if Chart not yet loaded; display code that will run once Chart is available
}

// Register a minimal CodeMirror mode for GCSL if CodeMirror is available.
// This mode highlights comments, strings, numbers, keywords and identifiers.
try {
    if (typeof CodeMirror !== 'undefined' && !CodeMirror.modes['gcsl']) {
        CodeMirror.defineMode('gcsl', function(config, parserConfig) {
            const keywords = new Set(['char','add','set','stats','target','energy','active','options','if','else','for','while','return','break','continue','let','fn','skill','burst','attack','dash','charge']);

            return {
                token: function(stream, state) {
                    if (stream.match('//') || stream.match('/*')) {
                        // line or block comment
                        if (stream.match('//')) {
                            stream.skipToEnd();
                            return 'comment';
                        }
                        // block comment start
                        while (!stream.eol()) {
                            if (stream.match('*/')) break;
                            stream.next();
                        }
                        return 'comment';
                    }

                    if (stream.match(/^(?:"(?:[^\\"]|\\.)*"|\'(?:[^\\']|\\.)*\')/)) {
                        return 'string';
                    }

                    if (stream.match(/^\d+(?:\.\d+)?/)) {
                        return 'number';
                    }

                    if (stream.match(/^[A-Za-z_][A-Za-z0-9_]*/)) {
                        const cur = stream.current();
                        if (keywords.has(cur)) return 'keyword';
                        return 'variable';
                    }

                    // operators / punctuation
                    stream.next();
                    return null;
                }
            };
        });
    }
} catch (e) {
    console.warn('CodeMirror GCSL mode registration failed', e);
}

document.addEventListener('DOMContentLoaded', function() {
    debugLog('[WebUI] Initializing...');
    // Editor setup: CodeMirror preferred; fallback to textarea
    const textarea = document.getElementById('config-editor');
    // Initialize CodeMirror with updated settings
    try {
        cmEditor = CodeMirror.fromTextArea(textarea, {
            mode: 'gcsl',
            lineNumbers: true,
            lineWrapping: true,
            theme: document.documentElement.getAttribute('data-theme') === 'dark' ? 'material' : 'default',
            tabSize: 2,
            indentWithTabs: false,
            autofocus: true
        });

        // Bind Ctrl/Cmd+Enter to run
        cmEditor.addKeyMap({ 'Ctrl-Enter': runSimulation, 'Cmd-Enter': runSimulation });

        // Keep textarea fallback in sync
        cmEditor.on('change', () => {
            textarea.value = cmEditor.getValue();
        });
        // Enforce visual sizing on CodeMirror after init
        const cmWrapper = cmEditor.getWrapperElement();
        if (cmWrapper) {
            cmWrapper.style.height = '720px';
            cmWrapper.style.fontSize = '13px';
        }
        // ensure inner scroller matches
        const scroller = cmWrapper.querySelector('.CodeMirror-scroll');
        if (scroller) scroller.style.height = '720px';
        // Refresh to apply sizing
        cmEditor.refresh();
    } catch (e) {
        console.warn('CodeMirror init failed, falling back to textarea', e);
        cmEditor = null;
    }
    // Apply saved theme if any
    const savedTheme = localStorage.getItem('gcsim_theme');
    if (savedTheme) {
        document.documentElement.setAttribute('data-theme', savedTheme);
    }
    // Theme toggle button
    const themeBtn = document.getElementById('theme-toggle');
    if (themeBtn) {
        themeBtn.addEventListener('click', () => {
            const cur = document.documentElement.getAttribute('data-theme');
            const next = cur === 'dark' ? '' : 'dark';
            if (next) {
                document.documentElement.setAttribute('data-theme', next);
                localStorage.setItem('gcsim_theme', next);
            } else {
                document.documentElement.removeAttribute('data-theme');
                localStorage.removeItem('gcsim_theme');
            }
        });
    }
    
    // Set default config - simpler version for reliable execution
    const defaultConfig = `nefer char lvl=90/90 cons=0 talent=9,9,9;
nefer add weapon="blackmarrowlantern" refine=5 lvl=90/90;
nefer add set="notsu" count=4;
nefer add stats hp=4780 atk=311 em=187 em=187 cd=0.622; #main
nefer add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*2 er=0.0551*2 em=19.82*4 cr=0.0331*12 cd=0.0662*10;

aino char lvl=90/90 cons=6 talent=9,9,9;
aino add weapon="flameforgedinsight" refine=5 lvl=90/90;
aino add set="ins" count=4;
aino add stats hp=3571 er=0.511 em=139 cr=0.249; #main
aino add stats def%=0.05208*2 def=16.5312*2 hp=213.31*2 hp%=0.041664*2 atk=13.8936*2 atk%=0.041664*2 er=0.046284*2 em=16.6488*2 cr=0.027804*7 cd=0.055608*9;

lauma char lvl=90/90 cons=0 talent=9,9,9;
lauma add weapon="etherlightspindlelute" refine=5 lvl=90/90;
lauma add set="sms" count=4;
lauma add stats hp=4780 atk=311 em=187 em=187 em=187; #main
lauma add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*2 er=0.0551*8 em=19.82*6 cr=0.0331*10 cd=0.0662*4;

nahida char lvl=90/90 cons=0 talent=9,9,9;
nahida add weapon="widsith" refine=3 lvl=90/90;
nahida add set="deepwood" count=4;
nahida add stats hp=4780 atk=311 em=187 dendro%=0.466 cr=0.311; #main
nahida add stats def%=0.062*2 def=19.68*2 hp=253.94*2 hp%=0.0496*2 atk=16.54*2 atk%=0.0496*2 er=0.0551*11 em=19.82*2 cr=0.0331*8 cd=0.0662*7;

options swap_delay=12 iteration=1000; 
target lvl=100 resist=0.1 radius=2 pos=2.1,1.5 hp=999999999; 
energy every interval=480,720 amount=1;

active nahida;


for let i=0; i<4; i=i+1 {
  nahida skill;
  if .nahida.burst.ready && .nahida.energymax {
	nahida burst;
  } else {
	nahida attack:2;
  }
  aino skill, burst;
  lauma skill;
  if .lauma.burst.ready && .lauma.energymax {
	lauma burst;
  } else {
	lauma attack:2;
  }
  nefer skill, charge, dash, charge, dash, charge;
  nahida attack, skill, charge, attack;
  nefer skill, charge, dash, charge, dash, charge;
  if .nefer.burst.ready && .nefer.energymax {
    nefer dash, burst;
  }
}
`;
    
    textarea.value = defaultConfig;
    if (cmEditor) cmEditor.setValue(defaultConfig);
    debugLog('[WebUI] Default config loaded');
    
    // If CodeMirror not available, attach keyboard shortcuts to textarea
    if (!cmEditor) {
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
    }

    // Syntax highlight: update highlight pre to mirror textarea with spans
    const highlightEl = document.getElementById('config-highlight');
    // Prism language registration disabled: use local scanner fallback for highlighting.
    // The previous complex inline RegExp caused parsing issues in some environments, so
    // we simply prefer the local highlighter implementation below. If Prism is available
    // and you want custom language support, add a safe registration script separately.
    function escapeHtml(s) {
        return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    function highlightGcsl(text) {
        if (!text) return '';
        const rules = [
            {regex: /(?:\/\*[\s\S]*?\*\/|\/\/.*?(?:\n|$))/y, cls: 'gcsl-comment'},
            {regex: /(?:\"(?:[^\\\"]|\\.)*\"|\'(?:[^\\\']|\\.)*\')/y, cls: 'gcsl-string'},
            {regex: /\b\d+(?:\.\d+)?\b/y, cls: 'gcsl-number'},
            {regex: /\b(?:char|add|set|stats|target|energy|active|options|if|else|for|while|return|break|continue|let|fn|skill|burst|attack|dash|charge)\b/g, cls: 'gcsl-keyword'},
            {regex: /\b([A-Za-z_][A-Za-z0-9_]*)/y, cls: 'gcsl-fn'},
            {regex: /[+\-*/=<>!:]+/y, cls: 'gcsl-operator'},
            {regex: /\s+/y, cls: null},
            {regex: /./y, cls: null}
        ];

        let pos = 0;
        let out = '';
        const src = text;

        while (pos < src.length) {
            let matched = false;
            for (let r of rules) {
                r.regex.lastIndex = pos;
                const m = r.regex.exec(src);
                if (m && m.index === pos) {
                    const tok = m[0];
                    const escaped = escapeHtml(tok);
                    if (r.cls) {
                        out += `<span class="${r.cls}">${escaped}</span>`;
                    } else {
                        out += escaped;
                    }
                    pos += tok.length;
                    matched = true;
                    break;
                }
            }
            if (!matched) {
                // shouldn't happen, but advance one char to avoid infinite loop
                out += escapeHtml(src[pos]);
                pos++;
            }
        }

        return out;
    }

    function updateHighlight() {
        const val = (typeof cmEditor !== 'undefined' && cmEditor) ? cmEditor.getValue() : textarea.value;
        if (highlightEl) {
            // If Prism is loaded and has highlightElement, use it on the <code> child
            const codeEl = highlightEl.querySelector('code');
            if (window.Prism && typeof window.Prism.highlightElement === 'function' && codeEl) {
                // Update raw text inside code element and let Prism process it
                codeEl.textContent = val + '\n';
                try {
                    window.Prism.highlightElement(codeEl);
                    // If Prism didn't produce any token spans (unknown language), fallback to local highlighter
                    if (!codeEl.querySelector('.token')) {
                        codeEl.innerHTML = highlightGcsl(val) + '\n';
                    }
                } catch (e) {
                    // Fallback to local highlighter if Prism fails
                    codeEl.innerHTML = highlightGcsl(val) + '\n';
                }
            } else {
                // No Prism: use existing scanner-based highlighter
                // put result into the code element to keep DOM structure consistent
                const code = highlightEl.querySelector('code');
                if (code) {
                    code.innerHTML = highlightGcsl(val) + '\n';
                } else {
                    highlightEl.innerHTML = highlightGcsl(val) + '\n';
                }
            }
        }
    }

    // If using textarea fallback, sync scroll and input
    if (!cmEditor) {
        textarea.addEventListener('scroll', () => {
            if (highlightEl) highlightEl.scrollTop = textarea.scrollTop;
        });

        // update on input
        textarea.addEventListener('input', updateHighlight);
    } else {
        cmEditor.on('change', updateHighlight);
        cmEditor.on('scroll', () => {
            if (highlightEl) highlightEl.scrollTop = cmEditor.getScrollInfo().top;
        });
    }
    // initial highlight
    updateHighlight();
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
    debugLog('[WebUI] Starting simulation...');
    const textarea = document.getElementById('config-editor');
    const config = (typeof cmEditor !== 'undefined' && cmEditor) ? cmEditor.getValue() : textarea.value;
    const errorMsg = document.getElementById('error-message');
    const loading = document.getElementById('loading');
    const resultsContainer = document.getElementById('results-container');
    const runButton = document.querySelector('.btn-run');
    
    debugLog('[WebUI] Config length:', config.length);
    
    // Hide previous results and errors
    errorMsg.style.display = 'none';
    resultsContainer.style.display = 'none';
    loading.style.display = 'block';
    runButton.disabled = true;
    clearErrorHighlights();
    
    try {
    debugLog('[WebUI] Sending request to /api/simulate');
        const response = await fetch('/api/simulate', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ config })
        });
        
    debugLog('[WebUI] Response status:', response.status);
        
        loading.style.display = 'none';
        runButton.disabled = false;
        
        if (!response.ok) {
            const error = await response.json();
            console.error('[WebUI] Error response:', error);
            handleError(error);
            return;
        }
        
        const result = await response.json();
    debugLog('[WebUI] Simulation result:', result);
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
    debugLog('[WebUI] Handling error:', error);
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
    debugLog('[WebUI] Displaying results...');
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
    
    debugLog('[WebUI] Results displayed successfully');
    
    // Scroll to results
    resultsContainer.scrollIntoView({ behavior: 'smooth' });
}

function displayStatistics(result) {
    debugLog('[WebUI] Displaying statistics...');
    const stats = result.statistics || {};
    
    // Extract main statistics with stdev
    const dps = stats.dps?.mean || 0;
    const dpsStd = stats.dps?.sd || 0;
    const eps = stats.eps?.mean || 0;
    const epsStd = stats.eps?.sd || 0;
    const rps = stats.rps?.mean || 0;
    const rpsStd = stats.rps?.sd || 0;
    const hps = stats.hps?.mean || 0;
    const hpsStd = stats.hps?.sd || 0;
    const shp = stats.shp?.mean || 0;
    const shpStd = stats.shp?.sd || 0;
    const duration = stats.duration?.mean || result.simulator_settings?.duration || 0;
    const durationStd = stats.duration?.sd || 0;
    
    debugLog('[WebUI] Stats:', { dps, eps, rps, hps, shp, duration });
    
    // Display with 2 decimal places and stdev
    document.getElementById('stat-dps').innerHTML = formatStatWithStdev(dps, dpsStd);
    document.getElementById('stat-eps').innerHTML = formatStatWithStdev(eps, epsStd);
    document.getElementById('stat-rps').innerHTML = formatStatWithStdev(rps, rpsStd);
    document.getElementById('stat-hps').innerHTML = formatStatWithStdev(hps, hpsStd);
    document.getElementById('stat-shp').innerHTML = formatStatWithStdev(shp, shpStd);
    document.getElementById('stat-dur').innerHTML = formatStatWithStdev(duration, durationStd);
}

function displayCharacters(result) {
    debugLog('[WebUI] Displaying characters...');
    debugLog('[WebUI] Full result keys:', Object.keys(result));
    const container = document.getElementById('characters-list');
    container.innerHTML = '';
    
    // Add grid wrapper
    const gridDiv = document.createElement('div');
    gridDiv.className = 'characters-grid';
    
    if (!result.character_details || result.character_details.length === 0) {
        container.innerHTML = '<p>キャラクター情報がありません</p>';
        return;
    }
    
    result.character_details.forEach((char, idx) => {
        console.log(`[WebUI] Character ${idx} keys:`, Object.keys(char));
        console.log(`[WebUI] Character ${idx} data:`, JSON.stringify(char, null, 2));
        
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
            setsHTML = `<div style="margin: 6px 0; font-size: 0.85rem;"><strong>聖遺物:</strong> ${setsList}</div>`;
        }
        
        // Stats display - use snapshot_stats for final values
        let statsHTML = '';
        const snapshotStats = char.snapshot_stats || char.snapshot || [];
        if (snapshotStats && snapshotStats.length > 0) {
            console.log(`[WebUI] Character ${name} snapshot_stats:`, snapshotStats);
            
            const finalHP = snapshotStats[3] || 0;
            const finalATK = snapshotStats[5] || 0;
            const finalDEF = snapshotStats[2] || 0;
            const finalEM = snapshotStats[8] || 0;
            const finalCR = snapshotStats[9] || 0;
            const finalCD = snapshotStats[10] || 0;
            const finalER = snapshotStats[7] || 0;
            
            const statDefs = [
                { name: 'HP', value: finalHP, format: (v) => Math.round(v) },
                { name: '攻撃力', value: finalATK, format: (v) => Math.round(v) },
                { name: '防御力', value: finalDEF, format: (v) => Math.round(v) },
                { name: '元素熟知', value: finalEM, format: (v) => Math.round(v) },
                { name: '会心率', value: finalCR, format: (v) => (v * 100).toFixed(1) + '%' },
                { name: '会心ダメージ', value: finalCD, format: (v) => (v * 100).toFixed(1) + '%' },
                { name: '元素チャージ効率', value: finalER, format: (v) => (v * 100).toFixed(1) + '%' },
            ];
            
            console.log(`[WebUI] ${name} final stats: HP=${Math.round(finalHP)}, ATK=${Math.round(finalATK)}, DEF=${Math.round(finalDEF)}, EM=${Math.round(finalEM)}, CR=${(finalCR*100).toFixed(1)}%, CD=${(finalCD*100).toFixed(1)}%, ER=${(finalER*100).toFixed(1)}%`);
            
            statsHTML = '<div class="char-stats-list"><div style="margin: 6px 0; font-size: 0.85rem;"><strong>ステータス詳細:</strong></div>';
            statDefs.forEach(({name, value, format}) => {
                if (value !== undefined && value !== 0) {
                    statsHTML += `<div class="info-row">
                        <span class="info-label">${name}</span>
                        <span class="info-value">${format(value)}</span>
                    </div>`;
                }
            });
            statsHTML += '</div>';
        } else {
            console.log('[WebUI] No snapshot_stats found for character:', name);
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
        
        gridDiv.appendChild(charDiv);
    });
    
    container.appendChild(gridDiv);

    // Append the target info block once under the characters list
    try {
        const targetsBlockHtml = buildTargetsHTML(result);
        if (targetsBlockHtml && targetsBlockHtml.trim().length > 0) {
            const targetsDiv = document.createElement('div');
            // Reuse card styles so visuals match character cards exactly
            targetsDiv.className = 'card';
            targetsDiv.style.marginTop = '12px';
            targetsDiv.innerHTML = `<div class="card-content"><span class="card-title">ターゲット情報</span>${targetsBlockHtml}</div>`;
            container.appendChild(targetsDiv);
        }
    } catch (e) { console.warn('[WebUI] Could not append targets under characters', e); }
}

function displayTargetInfo(result) {
    const container = document.getElementById('target-details');
    // If the target info tab/element was removed from the DOM, skip rendering.
    if (!container) return;
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

// Build HTML block for all targets so it can be embedded under each character card
function buildTargetsHTML(result) {
    if (!result.target_details || result.target_details.length === 0) return '';
    let html = '<div style="margin-top:10px;"><strong>ターゲット情報:</strong>';
    result.target_details.forEach((target, idx) => {
        const name = target.name || `ターゲット ${idx + 1}`;
        const level = target.level || 1;
        const hp = target.hp || 0;
        let resistHTML = '';
        if (target.resist && Object.keys(target.resist).length > 0) {
            resistHTML = '<div style="margin-top:6px;">';
            for (const [element, resist] of Object.entries(target.resist)) {
                resistHTML += `<div class="info-row"><span class="info-label">${element}</span><span class="info-value">${(resist * 100).toFixed(1)}%</span></div>`;
            }
            resistHTML += '</div>';
        }
        html += `<div style="margin-top:8px; padding:8px; border:1px solid var(--muted-border); border-radius:6px; background:var(--card-bg);">
            <div class="info-row"><span class="info-label">${name}</span><span class="info-value">Lv.${level}</span></div>
            <div class="info-row"><span class="info-label">HP</span><span class="info-value">${formatNumber(hp)}</span></div>
            ${resistHTML}
        </div>`;
    });
    html += '</div>';
    return html;
}

function displayCharts(result) {
    console.log('[WebUI] Displaying charts...');
    console.log('[WebUI] Result structure:', Object.keys(result));
    console.log('[WebUI] Statistics:', result.statistics);
    
    // Destroy existing charts
    Object.values(charts).forEach(chart => {
        if (chart && typeof chart.destroy === 'function') chart.destroy();
    });
    charts = {};
    
    const stats = result.statistics || {};
    
    
    // Insert or update a raw statistics dump to help debugging field shapes
    try {
        let rawPanel = document.getElementById('raw-stats-panel');
        if (!rawPanel) {
            rawPanel = document.createElement('details');
            rawPanel.id = 'raw-stats-panel';
            rawPanel.style.margin = '10px 0';
            const summary = document.createElement('summary');
            summary.textContent = 'Raw statistics JSON (debug)';
            rawPanel.appendChild(summary);
            const pre = document.createElement('pre');
            pre.id = 'raw-stats-pre';
            pre.style.maxHeight = '300px';
            pre.style.overflow = 'auto';
            pre.style.background = 'var(--card-bg)';
            pre.style.border = '1px solid var(--muted-border)';
            pre.style.padding = '8px';
            rawPanel.appendChild(pre);
            resultsContainer.insertBefore(rawPanel, resultsContainer.firstChild);
        }
        const preEl = document.getElementById('raw-stats-pre');
        if (preEl) preEl.textContent = JSON.stringify(result.statistics || {}, null, 2);
    } catch (e) { console.warn('[WebUI] Could not render raw stats panel', e); }

    // Character DPS Chart (100% Stacked Bar Chart)
    if (result.character_details && result.character_details.length > 0) {
        const canvas = document.getElementById('char-dps-chart');
        if (!canvas) {
            console.error('[WebUI] Canvas element char-dps-chart not found');
        } else {
            console.log('[WebUI] Found canvas element:', canvas);
            const ctx = canvas.getContext('2d');
            console.log('[WebUI] Got 2d context:', ctx);
            
            // Build arrays for characters and their DPS, then compute a canonical ordering
            const charNames = [];
            const charDpsData = [];
            const charDpsSd = [];

            result.character_details.forEach((char, idx) => {
                const name = char.name || `キャラ${idx+1}`;
                charNames.push(name);
                // Try multiple possible locations for character DPS data
                let dpsValue = 0;
                let sdValue = 0;
                if (stats.character_dps && Array.isArray(stats.character_dps)) {
                    dpsValue = stats.character_dps[idx]?.mean || 0;
                    sdValue = (typeof stats.character_dps[idx]?.sd !== 'undefined') ? stats.character_dps[idx].sd : 0;
                } else if (stats.character_dps && typeof stats.character_dps === 'object') {
                    dpsValue = stats.character_dps[name]?.mean || 0;
                    sdValue = (typeof stats.character_dps[name]?.sd !== 'undefined') ? stats.character_dps[name].sd : 0;
                }
                charDpsData.push(dpsValue);
                charDpsSd.push(sdValue);
            });
            
            console.log('[WebUI] Character DPS data:', charNames, charDpsData);
            
                if (charDpsData.length > 0 && charDpsData.some(v => v > 0)) {
                    // Compute canonical ordering by DPS descending so other charts can follow the same order
                    const order = charDpsData.map((v, i) => ({ idx: i, dps: v }));
                    order.sort((a,b) => b.dps - a.dps);
                    const orderedCharNames = order.map(o => charNames[o.idx]);
                    const orderedCharDps = order.map(o => charDpsData[o.idx]);
                    const orderedCharSd = order.map(o => charDpsSd[o.idx]);
                    // store canonical ordering on the stats object for use by other charts
                    stats.__char_order = { order, orderedCharNames, orderedCharDps, orderedCharSd };
                    // pass an empty labels array so no 'チーム' label appears on the axis
                    charts.charDps = createStackedBarChart(ctx, [''], [orderedCharNames, orderedCharDps, orderedCharSd], 'キャラクター別DPS');
                } else {
                console.log('[WebUI] No character DPS data to display');
            }
        }
    }
    
    // Source DPS Chart
    const canvas2 = document.getElementById('source-dps-chart');
    if (!canvas2) {
        console.error('[WebUI] Canvas element source-dps-chart not found');
    } else {
        const ctx2 = canvas2.getContext('2d');
        let sourceData = {};



        // Try to extract source DPS data (flatten nested objects and sum values)
        const durationMean = (stats.duration && (typeof stats.duration.mean === 'number')) ? stats.duration.mean : (result.simulator_settings && result.simulator_settings.duration) || 0;
        if (stats.dps_by_element && Array.isArray(stats.dps_by_element)) {
            stats.dps_by_element.forEach((charData, idx) => {
                const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
                if (charData && typeof charData === 'object') {
                    Object.entries(charData).forEach(([element, data]) => {
                        const key = `${charName} (${element})`;
                        const mean = extractDescriptiveMean(data);
                        if (typeof mean === 'number' && mean > 0) sourceData[key] = mean;
                    });
                }
            });
        } else if (stats.source_dps && Array.isArray(stats.source_dps)) {
            // Prefer explicit SourceDps if provided: array per-character SourceStats
            stats.source_dps.forEach((sa, idx) => {
                const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
                if (sa && sa.sources) {
                    Object.entries(sa.sources).forEach(([source, ds]) => {
                        // ds may be a DescriptiveStats or numeric; extract intelligently
                        const mean = extractDescriptiveMean(ds);
                        const num = (mean !== null) ? mean : extractNumber(ds);
                        if (typeof num === 'number' && num > 0) sourceData[`${charName}: ${source}`] = num;
                    });
                }
            });
        } else {
            // Note: stats.source_damage_instances often contains raw instance counts rather than DPS.
            // To avoid plotting count-only data as DPS, we skip using source_damage_instances as a fallback.
            if (stats.source_damage_instances) console.log('[WebUI] source_damage_instances present but ignored (counts only)');
        }
        
        console.log('[WebUI] Source DPS data:', sourceData);
        
    const data = extractChartData(sourceData);
    // Prefer stats.source_dps (per-character SourceStats) for per-character ability DPS
    if (stats.source_dps && Array.isArray(stats.source_dps) && stats.source_dps.length > 0) {
    // Use the canonical ordering computed from character DPS if available so colors/order match
    const charNamesRaw = (result.character_details && Array.isArray(result.character_details)) ? result.character_details.map(c => c.name) : stats.source_dps.map((_,i) => `キャラ${i+1}`);
    const charNames = (stats.__char_order && stats.__char_order.orderedCharNames) ? stats.__char_order.orderedCharNames : charNamesRaw;
        // Collect ability/source keys from source_dps
        const abilitySet = new Set();
        stats.source_dps.forEach(sa => { if (sa && sa.sources) Object.keys(sa.sources).forEach(k => abilitySet.add(k)); });
        const abilities = Array.from(abilitySet);

        if (abilities.length > 0) {
            // Create a matrix matching sorted charNames order. source_dps is indexed by original character index,
            // so we need to map canonical ordering indices back to original indices in source_dps.
            const originalCharNames = (result.character_details && Array.isArray(result.character_details)) ? result.character_details.map(c => c.name) : stats.source_dps.map((_,i) => `キャラ${i+1}`);
            // Build a mapping from canonical position -> original index
            const canonicalToOriginal = [];
            if (stats.__char_order && stats.__char_order.order) {
                // order: array of {idx, dps} where idx is original index
                stats.__char_order.order.forEach(o => canonicalToOriginal.push(o.idx));
            } else {
                // default mapping: identity
                for (let i = 0; i < originalCharNames.length; i++) canonicalToOriginal.push(i);
            }

            const matrix = abilities.map(() => Array(canonicalToOriginal.length).fill(0));
            const metaMatrix = abilities.map(() => Array(canonicalToOriginal.length).fill(null));

            abilities.forEach((ability, aIdx) => {
                canonicalToOriginal.forEach((origIdx, cCanonicalIdx) => {
                    const sa = stats.source_dps[origIdx];
                    if (!sa || !sa.sources) return;
                    const ds = sa.sources[ability];
                    if (!ds) return;
                    const mean = (typeof ds.mean === 'number') ? ds.mean : 0;
                    const sd = (typeof ds.sd === 'number') ? ds.sd : 0;
                    const min = (typeof ds.min === 'number') ? ds.min : 0;
                    const max = (typeof ds.max === 'number') ? ds.max : 0;
                    matrix[aIdx][cCanonicalIdx] = mean;
                    metaMatrix[aIdx][cCanonicalIdx] = { mean, sd, min, max };
                });
            });

            // Request the abilities chart use a slightly thicker bar and 5px vertical gap
            charts.sourceDps = createStackedAbilitiesChart(ctx2, charNames, abilities, matrix, 'キャラクター別 能力DPS', metaMatrix, { barThickness: 24, verticalPadding: 5 });
        } else if (data.labels.length > 0) {
            charts.sourceDps = createBarChart(ctx2, data.labels, data.values, 'ソース別DPS');
        } else {
            console.log('[WebUI] No source DPS data to display');
            try { showEmptyChartPlaceholder(ctx2.canvas.parentElement, 'ソース別DPS のデータがありません'); } catch(e) {}
        }
    } else if (stats.character_actions && Array.isArray(stats.character_actions) && stats.character_actions.length > 0) {
        // character_actions usually contains action counts (not DPS). Do not use it for DPS plotting.
        console.log('[WebUI] character_actions present but ignored for DPS (contains counts)');
        if (data.labels.length > 0) charts.sourceDps = createBarChart(ctx2, data.labels, data.values, 'ソース別DPS');
        else console.log('[WebUI] No source DPS data to display');
    } else {
        if (data.labels.length > 0) charts.sourceDps = createBarChart(ctx2, data.labels, data.values, 'ソース別DPS');
        else console.log('[WebUI] No source DPS data to display');
    }
    }
    
    // Damage Distribution Chart (Time-based line chart)
    const canvas3 = document.getElementById('damage-dist-chart');
    if (!canvas3) {
        console.error('[WebUI] Canvas element damage-dist-chart not found');
    } else {
        const ctx3 = canvas3.getContext('2d');
        if (stats.damage_buckets) {
        const buckets = stats.damage_buckets;
        const bucketSize = buckets.bucket_size || 30; // bucket size in frames
        const bucketData = buckets.buckets || [];

        // Convert frame-based bucket indices to seconds. 1s = 60 frames.
        const timeLabels = bucketData.map((_, idx) => {
            const frames = idx * bucketSize;
            const secs = frames / 60;
            // show integer seconds when >=1s, otherwise show 2 decimals
            return secs >= 1 ? `${secs.toFixed(0)}s` : `${secs.toFixed(2)}s`;
        });
        const damageValues = bucketData.map(bucket => bucket?.mean || 0);
        
        console.log('[WebUI] Damage distribution data:', timeLabels.length, 'buckets');
        
        if (timeLabels.length > 0) {
            // Render distribution with much larger vertical footprint per user request
            charts.damageDist = createLineChart(ctx3, timeLabels, damageValues, 'ダメージ', { heightPx: 480 });
        }
    } else {
        console.log('[WebUI] No damage distribution data');
    }
    }
    
    // Energy chart removed (per request)
    
    // Reaction Count Chart: show per-reaction bars where each bar is stacked by character counts
    (function() {
        const canvas5 = document.getElementById('reaction-count-chart');
        if (!canvas5) {
            console.error('[WebUI] Canvas element reaction-count-chart not found');
            return;
        }
        const ctx5 = canvas5.getContext('2d');
        if (!(stats.source_reactions && Array.isArray(stats.source_reactions))) {
            try { showEmptyChartPlaceholder(ctx5.canvas.parentElement, '反応回数のデータがありません'); } catch(e) {}
            return;
        }

        // Collect reaction types and per-character counts
        const reactionsSet = new Set();
        const charNames = [];
        const perCharReactions = []; // array of maps { reaction -> count }

        stats.source_reactions.forEach((charReactions, idx) => {
            const charName = result.character_details?.[idx]?.name || `キャラ${idx+1}`;
            charNames.push(charName);
            const map = {};
            if (charReactions && typeof charReactions === 'object') {
                Object.entries(charReactions).forEach(([reaction, rawVal]) => {
                    const num = extractNumber(rawVal) || 0;
                    if (num !== 0) {
                        map[reaction] = num;
                        reactionsSet.add(reaction);
                    }
                });
            }
            perCharReactions.push(map);
        });

        const reactions = Array.from(reactionsSet);
        if (reactions.length === 0) {
            try { showEmptyChartPlaceholder(ctx5.canvas.parentElement, '反応回数のデータがありません'); } catch(e) {}
            return;
        }

        // Build datasets: one dataset per character so each character has a consistent color across reaction bars
        const datasets = charNames.map((cn, ci) => {
            const data = reactions.map(r => perCharReactions[ci][r] || 0);
            return {
                label: cn,
                data,
                backgroundColor: getCharColor(ci),
                stack: 'Stack 0',
            };
        });

        // Ensure container height and create stacked bar chart (vertical categories = reactions)
        ensureContainerHeight(ctx5, Math.max(200, reactions.length * 30));
        charts.reactions = new Chart(ctx5, {
            type: 'bar',
            data: { labels: reactions, datasets },
            options: {
                // Horizontal bars: categories on Y axis
                indexAxis: 'y',
                plugins: { 
                    legend: { position: 'top' },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                // Reaction chart should show raw counts (not percents).
                                const raw = ctxRawValue(context);
                                return `${context.dataset.label || ''}: ${formatValue(raw, 2)}`;
                            }
                        }
                    }
                },
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: { stacked: true, beginAtZero: true },
                    y: { stacked: true }
                }
            }
        });
        try { scheduleChartResize(charts.reactions, ctx5); } catch(e) {}
    })();

    // Aura Uptime Chart: per-target stacked bars where each element is a segment.
    (function() {
        const canvas6 = document.getElementById('aura-uptime-chart');
        if (!canvas6) {
            console.error('[WebUI] Canvas element aura-uptime-chart not found');
            return;
        }
        const ctx6 = canvas6.getContext('2d');
        if (!(stats.target_aura_uptime && Array.isArray(stats.target_aura_uptime))) {
            try { showEmptyChartPlaceholder(ctx6.canvas.parentElement, '付着時間のデータがありません'); } catch(e) {}
            return;
        }

        // Each entry in target_aura_uptime represents a target: map of element->value (0..10000)
        const targetLabels = [];
        const elementSet = new Set();
        const perTarget = []; // array of maps element->value

        stats.target_aura_uptime.forEach((targetAura, tidx) => {
            const label = `ターゲット${tidx+1}`;
            targetLabels.push(label);
            const map = {};
            if (targetAura && typeof targetAura === 'object') {
                // The proto defines TargetAuraUptime as []*SourceStats, where SourceStats
                // has a `sources` map of element->DescriptiveStats. Some serializations
                // therefore nest elements under `sources`. Prefer that shape; fall back
                // to treating top-level keys as element names if `sources` isn't present.
                const inner = (targetAura.sources && typeof targetAura.sources === 'object') ? targetAura.sources : targetAura;
                Object.entries(inner).forEach(([element, rawVal]) => {
                    // value may be a numeric or a DescriptiveStats-like object
                    const num = extractNumber(rawVal) || 0;
                    if (num !== 0) {
                        // clamp between 0 and 10000
                        const clamped = Math.max(0, Math.min(10000, num));
                        map[element] = clamped;
                        elementSet.add(element);
                    }
                });
            }
            perTarget.push(map);
        });

        const elements = Array.from(elementSet);
        if (elements.length === 0) {
            try { showEmptyChartPlaceholder(ctx6.canvas.parentElement, '付着時間のデータがありません'); } catch(e) {}
            return;
        }

        // Detect numeric scale and convert to percent robustly. Server may return:
        // - fractions (0..1),
        // - percentages (0..100), or
        // - scaled integers (0..10000) as earlier code assumed.
        let globalMax = 0;
        perTarget.forEach(pt => {
            Object.values(pt).forEach(v => {
                if (typeof v === 'number' && Number.isFinite(v)) globalMax = Math.max(globalMax, Math.abs(v));
            });
        });

        const toPercent = (v) => {
            if (!v || !Number.isFinite(v)) return 0;
            // if values look like fractions (<= 1.01) -> multiply by 100
            if (globalMax <= 1.01) return v * 100;
            // if values look like percents already (<= 100.5) -> leave as-is
            if (globalMax <= 100.5) return v;
            // if values look like 0..10000 scale -> convert
            if (globalMax <= 10000) return v / 10000 * 100;
            // fallback: clamp to 0..100
            return Math.max(0, Math.min(100, v));
        };

        // Build datasets: one dataset per element, data is per-target percent values
        const datasets = elements.map((el, ei) => {
            const data = perTarget.map(pt => toPercent(pt[el] || 0));
            // color scheme: derive from element name via hash -> hue
            const hue = Math.abs(hashCode(el)) % 360;
            return {
                label: el,
                data,
                backgroundColor: `hsl(${hue}deg 70% 50%)`,
                stack: 'Stack 0',
            };
        });

        ensureContainerHeight(ctx6, Math.max(200, targetLabels.length * 50));
        charts.aura = new Chart(ctx6, {
            type: 'bar',
            data: { labels: targetLabels, datasets },
            options: {
                // Horizontal bars: categories (targets) on Y axis, percent on X axis
                indexAxis: 'y',
                plugins: { 
                    legend: { position: 'top' },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                // Aura uptime values are already converted to percent by toPercent
                                const raw = ctxRawValue(context);
                                return `${context.dataset.label || ''}: ${formatValue(raw, 2, '%')}`;
                            }
                        }
                    }
                },
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: { stacked: true, beginAtZero: true, max: 100, ticks: { callback: function(v){ return v + '%'; } } },
                    y: { stacked: true, grid: { display: false } }
                }
            }
        });
        try { scheduleChartResize(charts.aura, ctx6); } catch(e) {}
    })();
    
    console.log('[WebUI] Charts displayed, active charts:', Object.keys(charts));
}

function extractChartData(dataObj) {
    const labels = [];
    const values = [];
    
    if (typeof dataObj === 'object' && dataObj !== null) {
        for (const [key, value] of Object.entries(dataObj)) {
            if (typeof value === 'number') {
                if (!Number.isFinite(value)) continue;
                labels.push(key);
                values.push(value);
            } else if (typeof value === 'object' && value !== null && typeof value.mean === 'number') {
                if (!Number.isFinite(value.mean)) continue;
                labels.push(key);
                values.push(value.mean);
            }
        }
    }
    
    return { labels, values };
}

function showEmptyChartPlaceholder(containerEl, text) {
    try {
        if (!containerEl) return;
        // remove any existing placeholder
        const existing = containerEl.querySelector('.chart-empty-placeholder');
        if (existing) existing.remove();
        const div = document.createElement('div');
        div.className = 'chart-empty-placeholder';
        div.style.padding = '24px';
        div.style.color = 'var(--muted)';
        div.style.fontSize = '0.95rem';
        div.style.textAlign = 'left';
        div.textContent = text || 'データがありません';
        containerEl.appendChild(div);
    } catch (e) { /* ignore */ }
}

function createStackedBarChart(ctx, categories, [charNames, charValues, charSd], title) {
    
    // Calculate percentages
    const total = charValues.reduce((a, b) => a + b, 0);
    const percentages = charValues.map(v => total > 0 ? (v / total) * 100 : 0);
    
    console.log('[WebUI] Calculated percentages:', percentages);
    
    // Use global character palette so colors are consistent across charts
    const palette = CHAR_PALETTE;

    const datasets = charNames.map((name, idx) => {
        const hex = palette[idx % palette.length];
        const bg = hexToRgba(hex, 0.85);
        const border = hexToRgba(hex, 1);
        return {
            label: name,
            data: [percentages[idx]],
            stack: 'stack1',
            backgroundColor: bg,
            borderColor: border,
            borderWidth: 1,
            hoverBackgroundColor: hexToRgba(hex, 0.95),
            // Make the bar thickness approximately 24px
            barThickness: 48,
            maxBarThickness: 48
            ,categoryPercentage: 1.0
            ,barPercentage: 1.0
        };
    });
    
    // Prepare datasets; avoid verbose debug logging in production
    // Compute desired visual height using bar thickness and category count, and set canvas size
    const numRows = (Array.isArray(categories) && categories.length > 0) ? categories.length : 1;
    const barThickness = 48;
    const verticalPadding = 6;
    const legendSpace = 20;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);

    const chart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: categories,
            datasets: datasets
        },
        options: {
            // Render horizontally: categories on the Y axis, values (percent) on the X axis
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            layout: { padding: { top: 0, bottom: 0, left: 0, right: 0 } },
            plugins: {
                legend: {
                    display: true,
                    position: 'bottom',
                    labels: { boxWidth: 12, padding: 4 }
                },
                title: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        title: function() { return ''; },
                        label: function(context) {
                            const charIdx = context.datasetIndex;
                            const dps = charValues[charIdx] || 0;
                            const sd = (charSd && typeof charSd[charIdx] !== 'undefined' && charSd[charIdx] !== null) ? charSd[charIdx] : null;
                            const pct = percentages[charIdx];
                            // Show percentage with 2 decimals, DPS with thousands separator, stdev with 2 decimals or 'n/a'
                            const pctStr = pct.toFixed(2) + '%';
                            const sdStr = (sd === null) ? 'n/a' : sd.toFixed(2);
                            // DPS with 2 decimal places and locale formatting
                            const dpsStr = Number(dps).toLocaleString('ja-JP', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                            return `${context.dataset.label}: ${pctStr} (DPS: ${dpsStr} ± ${sdStr})`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    stacked: true,
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: function(value) { return value + '%'; },
                        padding: 4
                    }
                ,grid: { drawBorder: false, display: false }
                },
                y: {
                    stacked: true,
                    display: false,
                    grid: { display: false }
                }
            }
        }
    });
    // force a resize/update in case Chart computed wrong initial size
    try { chart.resize(); chart.update(); } catch (e) { /* ignore */ }
    // No data table is added beneath the chart (user requested removal)
    
    console.log('[WebUI] Chart created successfully, returning chart object');
    
    return chart;
}

// Ensure parent/ancestor container reserves a given height (px) to prevent overlapping charts
function ensureContainerHeight(ctx, desiredHeightPx) {
    try {
        const canvas = ctx && ctx.canvas ? ctx.canvas : null;
        if (!canvas) return;
        const parent = canvas.parentElement;
        if (parent) {
            parent.style.setProperty('min-height', Math.max(120, desiredHeightPx) + 'px', 'important');
        }
        // ensure ancestor columns (common .col) also have some reserved height
        let el = parent;
        let depth = 0;
        while (el && depth < 4) {
            if (el.classList && el.classList.contains('col')) {
                el.style.setProperty('min-height', Math.max(120, desiredHeightPx) + 'px', 'important');
                break;
            }
            el = el.parentElement;
            depth++;
        }
    } catch (e) { /* ignore */ }
}

// When a chart is created while its container is hidden, Chart.js may compute sizes as 0.
// Retry resize/update a few times until the canvas has a non-zero width.
function scheduleChartResize(chart, ctx, maxAttempts = 8) {
    try {
        let attempts = 0;
        const tryResize = () => {
            attempts++;
            const w = (ctx && ctx.canvas) ? ctx.canvas.offsetWidth : 0;
            if (w > 0 || attempts >= maxAttempts) {
                try { if (chart && typeof chart.resize === 'function') { chart.resize(); chart.update(); } } catch(e) {}
            } else {
                setTimeout(tryResize, 120);
            }
        };
        setTimeout(tryResize, 120);
    } catch (e) { /* ignore */ }
}

function createBarChart(ctx, labels, data, label, meta) {
    // Use global palette and compute simple colors
    const palette = CHAR_PALETTE;
    const bgColors = labels.map((_, i) => hexToRgba(palette[i % palette.length], 0.75));
    const borderColors = labels.map((_, i) => hexToRgba(palette[i % palette.length], 1));

    // Compute desired canvas height and set buffer via helper
    const numRows = labels.length || 1;
    const barThickness = 48;
    const verticalPadding = 6;
    const legendSpace = 8;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);

    const datasets = [{
        label: label,
        data: data,
        backgroundColor: bgColors,
        borderColor: borderColors,
        borderWidth: 1,
        barThickness: 48,
        maxBarThickness: 48,
        categoryPercentage: 1.0,
        barPercentage: 0.9
    }];

    const chart = new Chart(ctx, {
        type: 'bar',
        data: { labels: labels, datasets: datasets },
        options: {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            layout: { padding: { top:0, bottom:0, left:0, right:0 } },
            plugins: { 
                legend: { display: false },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const idx = context.dataIndex;
                            const val = data[idx] || 0;
                            // if meta provided and has descriptive stats for this label, show them
                            if (meta && meta[idx]) {
                                const m = meta[idx];
                                const mean = (typeof m.mean === 'number') ? m.mean : val;
                                const sd = (typeof m.sd === 'number') ? m.sd : null;
                                const min = (typeof m.min === 'number') ? m.min : null;
                                const max = (typeof m.max === 'number') ? m.max : null;
                                const sdStr = sd === null ? 'n/a' : sd.toFixed(2);
                                return `${context.label}: ${mean.toFixed(2)} ± ${sdStr}`;
                            }
                            return `${context.label}: ${typeof val === 'number' ? val.toFixed(2) : val}`;
                        }
                    }
                }
            },
            scales: {
                x: { beginAtZero: true, grid: { display: false }, ticks: { padding: 4 } },
                y: { grid: { display: false }, ticks: { padding: 6 } }
            }
        }
    });
    try { chart.resize(); chart.update(); } catch (e) { /* ignore */ }
    // No data table is added beneath the chart (user requested removal)
    
    return chart;
}

function createLineChart(ctx, labels, data, label, options) {
    // options.heightPx: desired visual height in pixels (default small for distribution)
    const opts = Object.assign({ heightPx: 140 }, options || {});

    // Set canvas visual size using helper to avoid duplication
    setCanvasVisualSize(ctx, opts.heightPx);

    const chart = new Chart(ctx, {
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
            maintainAspectRatio: false,
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
    try { chart.resize(); chart.update(); } catch (e) { /* ignore */ }
    // No data table is added beneath the chart (user requested removal)

    return chart;
}

// Create a horizontal stacked bar chart where each stack segment is an ability and bars are characters
function createStackedAbilitiesChart(ctx, charNames, abilities, matrix, title, metaMatrix, options) {
    // abilities: array of ability/source labels (bars)
    // matrix: abilities.length x charNames.length numeric matrix
    // metaMatrix: abilities.length x charNames.length metadata

    // Sort abilities by total DPS descending
    const totalByAbility = abilities.map((ab, aIdx) => ({ idx: aIdx, total: matrix[aIdx].reduce((s,v) => s + (v||0), 0) }));
    totalByAbility.sort((a,b) => b.total - a.total);
    const sortedAbilities = totalByAbility.map(t => abilities[t.idx]);
    const sortedMatrix = totalByAbility.map(t => matrix[t.idx]);
    const sortedMeta = metaMatrix ? totalByAbility.map(t => metaMatrix[t.idx]) : null;

    // Options with sensible defaults
    const opts = Object.assign({ barThickness: 18, verticalPadding: 6 }, options || {});

    // Build datasets per character so each stack segment uses character color
    const datasets = charNames.map((char, cIdx) => {
        const hex = getCharColor(cIdx);
        const bg = hexToRgba(hex, 0.85);
        const border = hexToRgba(hex, 1);
        // extract values across sorted abilities
        const data = sortedMatrix.map(row => row[cIdx] || 0);
        return {
            label: char,
            data: data,
            stack: 'stack1',
            backgroundColor: bg,
            borderColor: border,
            borderWidth: 1,
            barThickness: opts.barThickness,
            maxBarThickness: opts.barThickness,
            categoryPercentage: 1.0,
            barPercentage: 1.0
        };
    });

    // compute desired height and set canvas size via helper
    const numRows = sortedAbilities.length || 1;
    const barThickness = opts.barThickness || 18;
    const verticalPadding = (typeof opts.verticalPadding === 'number') ? opts.verticalPadding : 6;
    const legendSpace = 8;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);

    const chart = new Chart(ctx, {
        type: 'bar',
        data: { labels: sortedAbilities, datasets: datasets },
        options: {
            indexAxis: 'y',
            responsive: true,
            maintainAspectRatio: false,
            layout: { padding: { top:0, bottom:0, left:0, right:0 } },
            plugins: {
                legend: { display: true, position: 'bottom', labels: { boxWidth: 12, padding: 4 } },
                tooltip: {
                    callbacks: {
                        title: function() { return ''; },
                        label: function(context) {
                            const charIdx = context.datasetIndex;
                            const abilityIdx = context.dataIndex;
                            const ability = context.chart.data.labels[abilityIdx];
                            const val = context.dataset.data[abilityIdx] || 0;
                            // metaMatrix is abilities x chars; ability order is sortedAbilities
                            if (sortedMeta && sortedMeta[abilityIdx] && sortedMeta[abilityIdx][charIdx]) {
                                const m = sortedMeta[abilityIdx][charIdx];
                                const lines = [];
                                lines.push(`${charNames[charIdx]}: ${ability}`);
                                lines.push(`mean ${m.mean.toFixed(2)}`);
                                lines.push(`min ${m.min.toFixed(2)}`);
                                lines.push(`max ${m.max.toFixed(2)}`);
                                lines.push(`std ${m.sd.toFixed(2)}`);
                                return lines;
                            }
                            return `${charNames[charIdx]}: ${ability}: ${typeof val === 'number' ? val.toFixed(2) : val}`;
                        }
                    }
                }
            },
            scales: { x: { stacked: true, grid: { display: false }, ticks: { padding: 4 } }, y: { stacked: true, grid: { display: false }, ticks: { padding: 6 } } }
        }
    });

    try { chart.resize(); chart.update(); } catch (e) { /* ignore */ }
    return chart;
}

function addChartLegend(ctx, labels, colors) {
    const container = (ctx && ctx.canvas && ctx.canvas.parentElement) ? ctx.canvas.parentElement : (ctx && ctx.parentElement) ? ctx.parentElement : null;
    if (!container) {
        console.warn('[WebUI] Chart container not found for addChartLegend');
        return;
    }
    let legendDiv = container.querySelector('.chart-legend');
    
    if (!legendDiv) {
        legendDiv = document.createElement('div');
        legendDiv.className = 'chart-legend';
        container.appendChild(legendDiv);
    }
    
    let html = '<div class="chart-legend-title">凡例</div>';
    labels.forEach((label, idx) => {
        const color = colors[idx % colors.length];
        html += `<div class="chart-legend-item">
            <div class="chart-legend-color" style="background-color: ${color};"></div>
            <span>${label}</span>
        </div>`;
    });
    
    legendDiv.innerHTML = html;
}

// chart data tables under canvases were intentionally removed per user request

function formatNumber(num) {
    if (num === undefined || num === null) return '-';
    if (typeof num !== 'number') return num;
    
    // Format without K/M suffixes
    return Math.round(num).toLocaleString('ja-JP');
}

function formatStatWithStdev(mean, stdev) {
    if (mean === undefined || mean === null) return '-';
    
    // Format mean with 2 decimal places
    const meanFormatted = mean.toFixed(2);
    
    // Format stdev with 2 decimal places
    const stdevFormatted = stdev ? stdev.toFixed(2) : '0.00';
    
    // Return HTML with main value and small stdev below
    return `${meanFormatted}<br><small style="font-size: 0.5em; font-weight: 400; color: #999;">±${stdevFormatted}</small>`;
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

// Make functions available globally for onclick handlers
window.runSimulation = runSimulation;
window.switchTab = switchTab;
