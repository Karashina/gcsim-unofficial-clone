// Initialize charts storage
let charts = {};
// Global reference to CodeMirror editor (if initialized)
var cmEditor = null;
// Global references for optimizer mode editors
var cmEditorOriginal = null;
var cmEditorOptimized = null;

// Screen navigation function
function setupScreenNavigation() {
    // Use event delegation on the parent container so dynamically replaced
    // tab buttons still work and we avoid needing to re-bind listeners.
    const tabsContainer = document.querySelector('.navbar-tabs');
    const screens = document.querySelectorAll('.screen');

    if (!tabsContainer) return;

    tabsContainer.addEventListener('click', function(e) {
        const btn = e.target.closest('.navbar-tab');
        if (!btn) return;

        // guard: ignore if a disabled button
        if (btn.disabled) return;

        const screenId = btn.getAttribute('data-screen');

        // Remove active class from all buttons in container
        tabsContainer.querySelectorAll('.navbar-tab').forEach(b => b.classList.remove('active'));
        // Add active class to clicked button
        btn.classList.add('active');

        // Hide all screens
        screens.forEach(screen => screen.classList.remove('active'));

        // Show selected screen
        const targetScreen = document.getElementById('screen-' + screenId);
        if (targetScreen) {
            targetScreen.classList.add('active');
            // If results screen is shown, ensure any lazy charts are rendered
            if (screenId === 'results' && typeof window.onResultsShown === 'function') {
                try { window.onResultsShown(); } catch (e) { console.warn('onResultsShown failed', e); }
            }
        }
    });
}

// Mode switching function for config manual input
function setupModeSwitch() {
    const modeButtons = document.querySelectorAll('.mode-btn');
    const modes = document.querySelectorAll('.config-mode');
    
    modeButtons.forEach(button => {
        button.addEventListener('click', function() {
            const modeId = this.getAttribute('data-mode');
            
            // Remove active class from all buttons
            modeButtons.forEach(btn => btn.classList.remove('active'));
            // Add active class to clicked button
            this.classList.add('active');
            
            // Hide all modes
            modes.forEach(mode => mode.classList.remove('active'));
            // Show selected mode
            const targetMode = document.getElementById('mode-' + modeId);
            if (targetMode) {
                targetMode.classList.add('active');
            }
        });
    });
}

// Function to run optimizer simulation
async function runOptimizerSimulation() {
    debugLog('[WebUI] Starting optimizer simulation...');
    const originalTextarea = document.getElementById('config-editor-original');
    const optimizedTextarea = document.getElementById('config-editor-optimized');
    const errorMsg = document.getElementById('error-message-optimizer');
    const loading = document.getElementById('loading-optimizer');
    const runButton = document.querySelector('#mode-optimizer .btn-run');
    
    // Get config from the appropriate editor
    let config = '';
    if (optimizedTextarea.value.trim()) {
        // Priority: use optimized config if it exists
        config = (typeof cmEditorOptimized !== 'undefined' && cmEditorOptimized) ? 
                 cmEditorOptimized.getValue() : optimizedTextarea.value;
    } else if (originalTextarea.value.trim()) {
        // Fallback: use original config
        config = (typeof cmEditorOriginal !== 'undefined' && cmEditorOriginal) ? 
                 cmEditorOriginal.getValue() : originalTextarea.value;
    }
    
    if (!config.trim()) {
        errorMsg.textContent = 'エラー: コンフィグを�?�力してください';
        errorMsg.style.display = 'block';
        return;
    }
    
    debugLog('[WebUI] Config length:', config.length);
    
    // Hide previous results and errors
    errorMsg.style.display = 'none';
    loading.style.display = 'block';
    runButton.disabled = true;
    
    try {
        debugLog('[WebUI] Sending request to /api/optimize');
        const response = await fetch('/api/optimize', {
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
            errorMsg.textContent = 'エラー: ' + (error.message || error.error || 'Optimizer実行に失敗しました');
            errorMsg.style.display = 'block';
            return;
        }
        
        const result = await response.json();
        debugLog('[WebUI] Optimizer result:', result);
        
        // Display optimized config in right panel
        if (result.optimized_config) {
            if (cmEditorOptimized) {
                cmEditorOptimized.setValue(result.optimized_config);
            } else {
                optimizedTextarea.value = result.optimized_config;
            }
        }
        
        // If results are available, display them
        if (result.statistics) {
            // Switch to results screen
            const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
            if (resultsTab) {
                resultsTab.click();
            }
            displayResults(result);
        }
        
    } catch (err) {
        console.error('[WebUI] Exception:', err);
        loading.style.display = 'none';
        runButton.disabled = false;
        errorMsg.textContent = 'エラー: ' + err.message;
        errorMsg.style.display = 'block';
    }
}

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

// --- Localization maps (from cmd/gcsim CSVs) ---------------------------------
// character name map: english key -> Japanese display name
// Build CHAR_TO_JP from authoritative CSV: cmd/gcsim/chatracterData/charactertoJP.csv
const CHAR_TO_JP = {
    "aino": "アイ�?",
    "aloy": "アーロイ",
    "itto": "荒瀧一�?",
    "alhaitham": "アルハイゼン",
    "albedo": "アルベド",
    "arlecchino": "アルレ�?キー�?",
    "amber": "アンバ�?�",
    "iansan": "イアンサ",
    "yelan": "夜蘭",
    "ineffa": "イネファ",
    "ifa": "イファ",
    "varesa": "ヴァレサ",
    "venti": "ウェン�?ィ",
    "yunjin": "雲菫",
    "eula": "エウルア",
    "escoffier": "エスコフィエ",
    "emilie": "エミリエ",
    "yanfei": "煙�?",
    "ororon": "オロルン",
    "kaveh": "カーヴェ",
    "kaeya": "ガイア",
    "kazuha": "楓原�?�?",
    "kachina": "カチ�?��?",
    "ayaka": "神里綾華",
    "ayato": "神里綾人",
    "gaming": "嘉�??",
    "ganyu": "甘雨",
    "xianyun": "閑雲",
    "kinich": "キィニチ",
    "candace": "キャン�?ィス",
    "ningguang": "凝�??",
    "kirara": "綺良�?",
    "kuki": "�?岐�?",
    "sara": "九条裟�?",
    "klee": "クレー",
    "clorinde": "クロリン�?",
    "keqing": "刻晴",
    "collei": "コレイ",
    "gorou": "ゴロー",
    "sayu": "早�?",
    "kokomi": "珊瑚宮�?海",
    "heizou": "鹿野院平蔵",
    "sigewinne": "シグウィン",
    "citlali": "シトラリ",
    "charlotte": "シャルロ�?�?",
    "xiangling": "香菱",
    "chevreuse": "シュヴルーズ",
    "xiao": "�?",
    "zhongli": "鍾離",
    "xilonen": "シロネン",
    "jean": "ジン",
    "xinyan": "辛炎",
    "shenhe": "申鶴",
    "skirk": "スカーク",
    "sucrose": "スクロース",
    "sethos": "セトス",
    "cyno": "セ�?",
    "dahlia": "ダリア",
    "tartaglia": "タルタリヤ",
    "chiori": "�?�?",
    "chasca": "チャスカ",
    "chongyun": "重雲",
    "diona": "�?ィオ�?",
    "dehya": "�?ィシア",
    "tighnari": "�?ィナリ",
    "diluc": "�?ィル�?ク",
    "thoma": "ト�?��?",
    "dori": "ドリー",
    "qiqi": "�?�?",
    "navia": "ナヴィア",
    "nahida": "ナヒーダ",
    "nefer" : "ネフェル",
    "nilou": "ニィロウ",
    "neuvillette": "ヌヴィレ�?�?",
    "noelle": "ノエル",
    "barbara": "バ�?�バラ",
    "baizhu": "白朮",
    "faruzan": "ファルザン",
    "fischl": "フィ�?シュル",
    "hutao": "胡�?",
    "furina": "フリー�?",
    "flins": "フリンズ",
    "freminet": "フレミネ",
    "bennett": "ベネ�?�?",
    "wanderer": "放浪�?",
    "beidou": "北斗",
    "mavuika": "マ�?�ヴィカ",
    "mika": "ミカ",
    "mualani": "�?アラ�?",
    "mona": "モ�?",
    "yaemiko": "八重神�?",
    "xingqiu": "行�?",
    "yumemizukimizuki": "夢見月瑞�?",
    "mizuki": "夢見月瑞�?",
    "yoimiya": "宵宮",
    "yaoyao": "ヨォーヨ",
    "raiden": "雷電�?�?",
    "lauma": "ラウ�?",
    "lanyan": "藍硯",
    "wriothesley": "リオセスリ",
    "lisa": "リサ",
    "lyney": "リ�?",
    "lynette": "リネッ�?",
    "layla": "レイラ",
    "razor": "レザー",
    "rosaria": "ロサリア",
    "lumineanemo": "�?(風)",
    "luminegeo": "�?(岩)",
    "lumineelectro": "�?(雷)",
    "luminedendro": "�?(�?)",
    "luminehydro": "�?(水)",
    "luminepyro": "�?(�?)",
    "luminecryo": "�?(氷)",
    "aethergeo": "空(岩)",
    "aetherelectro": "空(雷)",
    "aetherdendro": "空(�?)",
    "aetherhydro": "空(水)",
    "aetherpyro": "空(�?)",
    "aethercryo": "空(氷)"
};

// weapon name map: weapon key -> Japanese name (subset loaded from csv)
const WEAPON_TO_JP = {
    "mistsplitterreforged": "霧�?の廻�?",
    "aquilafavonia": "風鷹剣",
    "summitshaper": "斬山の�?",
    "skywardblade": "天空の�?",
    "freedomsworn": "蒼古なる�?�由への誓い",
    "primordialjadecutter": "磐岩結�?",
    "harangeppakufutsu": "波乱月白経津",
    "keyofkhajnisut": "聖顕�?�鍵",
    "lightoffoliarincision": "�?光�?�裁葉",
    "splendoroftranquilwaters": "静水流転の輝き",
    "uraku": "有楽御簾�?",
    "absolution": "赦罪",
    "peakpatrolsong": "岩峰を巡る�?",
    "theflute": "笛�?�剣",
    "theblacksword": "黒剣",
    "thealleyflash": "ダークアレイの�?�?",
    "swordofdescension": "降�?�の剣",
    "sacrificialsword": "祭礼の剣",
    "royallongsword": "旧貴族長剣",
    "prototyperancour": "斬岩・試�?",
    "amenomakageuchi": "天目影�?",
    "lionsroar": "匣中龍吟",
    "ironsting": "�?蜂�?�刺�?",
    "festeringdesire": "腐植�?�剣",
    "favoniussword": "西風剣",
    "cinnabarspindle": "シナバースピンドル",
    "blackclifflongsword": "黒岩の長剣",
    "sapwoodblade": "原木刀",
    "xiphosmoonlight": "サイフォスの月�?�か�?",
    "kagotsurubeisshin": "�?釣瓶一�?",
    "wolffang": "狼�?",
    "finaleofthedeep": "海淵のフィナ�?�レ",
    "moonweaversdawn": "月紡ぎ�?�曙�??",
    "harbingerofdawn": "黎�?��?�神剣",
    "darkironsword": "暗鉄剣",
    "travelershandysword": "�?道�?�剣",
    "fluteofezpitzal": "エズピツァルの�?",
    "calamityofeshu": "�?水の災�?",
    "serenityscall": "静謐�?��?",
    "filletblade": "チ虎魚�?�刀",
    "skyridersword": "飛天御剣",
    "coolsteel": "冷�?",
    "toukaboushigure": "東花坊時雨",
    "fleuvecendreferryman": "サーンドルの渡し�?",
    "dockhand": "船�?剣",

    // -- bows (weaponData/bow.csv) --
    "polarstar": "冬極の白�?",
    "thunderingpulse": "飛雷の鳴弦",
    "elegyfortheend": "終焉を�??く詩",
    "skywardharp": "天空の翼",
    "amosbow": "アモスの�?",
    "hunterspath": "狩人の�?",
    "aquasimulacra": "若水",
    "thefirstgreatmagic": "始まり�?�大魔�?",
    "heartstrings": "白雨�?弦",
    "astralvulturescrimsonplumage": "星鷲の�?き羽",
    "alleyhunter": "ダークアレイの狩人",
    "theviridescenthunt": "蒼�?の狩猟�?",
    "thestringless": "絶弦",
    "sacrificialbow": "祭礼の�?",
    "rust": "弓蔵",
    "royalbow": "旧貴族長�?",
    "prototypecrescent": "澹月�?�試�?",
    "predator": "プレ�?ター",
    "mouunsmoon": "曚雲の�?",
    "mitternachtswaltz": "幽夜�?�ワル�?",
    "hamayumi": "破魔�?��?",
    "favoniuswarbow": "西風猟�?",
    "compoundbow": "リングボウ",
    "blackcliffwarbow": "黒岩の戦�?",
    "windblumeode": "風花の頌�?",
    "endoftheline": "竭沢",
    "fadingtwilight": "落�?",
    "kingssquire": "王�?�近�?",
    "ibispiercer": "トキの嘴",
    "scionoftheblazingsun": "烈日の後嗣",
    "songofstillness": "静寂�?��?",
    "cloudforged": "築雲",
    "chainbreaker": "チェーンブレイカー",
    "flowerwreathedfeathers": "花飾り�?�羽",
    "snarehook": "�?網の�?",
    "ravenbow": "鴉羽の�?",
    "recurvebow": "リカーブ�?�ウ",
    "messenger": "�?使�?",
    "sharpshootersoath": "シャープシューターの誓い",
    "slingshot": "弾�?",

    // -- polearms (weaponData/polearm.csv) --
    "engulfinglightning": "草薙の稲�?",
    "skywardspine": "天空の�?",
    "pjws": "和璞鳶",
    "calamityqueller": "息災",
    "staffofhoma": "護摩の�?",
    "vortexvanquisher": "破天の�?",
    "staffofthescarletsands": "赤砂�?��?",
    "crimsonmoonssemblance": "赤月�?�シルエ�?�?",
    "lumidouceelegy": "ルミドゥースの挽�?",
    "fracturedhalo": "砕け散る�?�輪",
    "bloodsoakedruins": "血染めの荒れ地",
    "prototypestarglitter": "星鎌・試�?",
    "lithicspear": "�?岩長�?",
    "kitaincrossspear": "喜多院十文字�?",
    "thecatch": "「漁獲�?",
    "favoniuslance": "西風長�?",
    "dragonspinespear": "ドラゴンスピア",
    "dragonsbane": "匣中�?�?",
    "deathmatch": "死闘�?��?",
    "crescentpike": "流月の�?",
    "blackcliffpole": "黒岩の突�?",
    "wavebreakersfin": "斬波のひれ長",
    "royalspear": "旧貴族猟�?",
    "moonpiercer": "�?ーンピアサー",
    "missivewindspear": "風信の�?",
    "balladofthefjords": "フィヨルド�?��?",
    "rightfulreward": "正義の報酬",
    "dialogues": "砂中の賢�?達�?�問�?",
    "footprintoftherainbow": "虹の行方",
    "tamayuratei": "玉響停�?�御噺",
    "prospectorsshovel": "金掘り�?�シャベル",
    "halberd": "鉾�?",
    "blacktassel": "黒纓�?",
    "whitetassel": "白纓�?",

    // -- claymores (weaponData/claymore.csv) --
    "wolfsgravestone": "狼の末路",
    "redhornstonethresher": "赤角石塵�?�?",
    "theunforged": "無工の剣",
    "songofbrokenpines": "松韻の響く�??",
    "skywardpride": "天空の傲",
    "beaconofthereedsea": "葦海の�?",
    "verdict": "裁断",
    "fangofthemountainking": "山の王�?�長�?",
    "athousandblazingsuns": "�?烈�?�日輪",
    "whiteblind": "白影の剣",
    "thebell": "鐘�?�剣",
    "snowtombedstarsilver": "雪葬の星銀",
    "serpentspine": "螭龍�?�剣",
    "sacrificialgreatsword": "祭礼の大剣",
    "blackcliffslasher": "黒岩の斬刀",
    "akuoumaru": "惡王丸",
    "rainslasher": "雨�?",
    "prototypearchaic": "古華・試�?",
    "luxurioussealord": "銜玉の海�?",
    "lithicblade": "�?岩古剣",
    "katsuragikirinagamasa": "桂木斬長正",
    "favoniusgreatsword": "西風大剣",
    "royalgreatsword": "旧貴族大剣",
    "forestregalia": "深林�?�レガリア",
    "makhairaaquamarine": "マカイラの水色",
    "mailedflower": "�?彩の花",
    "talkingstick": "話死合い�?",
    "tidalshadow": "タイダル・シャド�?�",
    "portablepowersaw": "携帯型チェーンソー",
    "ultimateoverlordsmegamagicsword": "「スーパ�?�アル�?ィメ�?ト�?王魔剣�?",
    "earthshaker": "アースシェイカー",
    "flameforgedinsight": "知恵の溶�?",
    "masterkey": "�?能の鍵",
    "skyridergreatsword": "飛天大御剣",
    "ferrousshadow": "�?影段平",
    "debateclub": "�?屈責�?",
    "whiteirongreatsword": "白�?の大剣",
     "bloodtaintedgreatsword": "龍血を浴びた剣",
     // -- catalyst (weaponData/catalyst.csv) --
     "memoryofdust": "浮世�?��?",
     "everlastingmoonglow": "不�?の月華",
     "skywardatlas": "天空の巻",
     "kagurasverity": "神楽の真意",
     "lostprayertothesacredwinds": "四風原�?�",
     "athousandfloatingdreams": "�?夜に浮か�?�夢",
     "tulaytullahsremembrance": "トゥライトゥーラの記�?�",
     "eternalflow": "�?�?流転の大典",
     "jadefallssplendor": "碧落の�?",
     "cashflowsupervision": "凛流�?�監視�?",
     "cranesechoingcall": "鶴鳴の余韻",
     "surfsup": "サーフィンタイ�?",
     "starcallerswatch": "祭星�?の眺�?",
     "sunnymorningsleepin": "寝正月�?�初晴",
     "nightweaverslookingglass": "夜を紡ぐ天鏡",
     "blackcliffagate": "黒岩の緋玉",
     "thewidsith": "流浪楽�?",
     "solarpearl": "匣中日�?",
     "sacrificialfragments": "祭礼の断�?",
     "royalgrimoire": "旧貴族秘法録",
     "prototypeamber": "金珀・試�?",
     "oathsworneye": "誓いの明瞳",
     "wineandsong": "ダークアレイの酒と詩",
     "mappamare": "�?国諸海の図�?",
     "hakushinring": "白辰の輪",
     "frostbearer": "冬忍�?�の�?",
     "favoniuscodex": "西風秘�?�",
     "eyeofperception": "昭�?",
     "dodocotales": "ドドコの物�?",
     "fruitoffulfillment": "満悦の�?",
     "wanderingevenstar": "彷徨える�?",
     "sacrificialjade": "古�?の�?",
     "flowingpurity": "純水流華",
     "ringofyaxche": "ヤシュチェの環",
     "ashgravendrinkinghorn": "蒼紋�?�角杯",
     "waveridingwhirl": "波乗りの旋回",
     "blackmarrowlantern": "烏�?の孤灯",
     "etherlightspindlelute": "天光�?�リュー�?",
     "magicguide": "魔導緒�?",
     "otherworldlystory": "異世界�?行�?",
     "emeraldorb": "翡玉法珠",
     "thrillingtalesofdragonslayers": "龍殺し�?�英傑�?",
     "twinnephrite": "特級�?�宝玉"
 };

// artifact name map (artifact set key -> Japanese name)
const ARTIFACT_TO_JP = {
    "maidenbeloved": "愛される少女",
    "songofdayspast": "在りし日の�?",
    "oceanhuedclam": "海染硨磲",
    "goldentroupe": "�?金�?��?団",
    "scrolloftheheroofcindercity": "灰燼の都に立つ英�?の絵巻",
    "fragmentofharmonicwhimsy": "諧律�?想の断�?",
    "vourukashasglow": "花海甘露の�?",
    "huskofopulentdreams": "華館夢醒形骸�?",
    "thunderingfury": "雷のような怒り",
    "thundersoother": "雷を鎮める尊�?",
    "noblesseoblige": "旧貴族�?�しつ�?",
    "instructor": "教�?",
    "gildeddreams": "金メ�?キの夢",
    "gladiatorsfinale": "剣闘士のフィナ�?�レ",
    "obsidiancodex": "黒曜の秘�?�",
    "retracingbolide": "�?飛�?�の流星",
    "desertpavilionchronicle": "砂上�?�楼閣の史話",
    "nighttimewhispersintheechoingwoods": "残響の森で囁かれる夜話",
    "vermillionhereafter": "辰砂往生録",
    "deepwoodmemories": "深林�?�記�?�",
    "finaleofthedeepgalleries": "深廊�?�終曲",
    "nymphsdream": "水仙�?�夢",
    "viridescentvenerer": "�?緑�?�影",
    "emblemofseveredfate": "絶縁�?�旗印",
    "tenacityofthemillelith": "�?岩牢固",
    "paleflame": "蒼白の�?",
    "wandererstroupe": "大地を流浪する楽団",
    "bloodstainedchivalry": "血染めの騎士�?",
    "heartofdepth": "沈淪の�?",
    "shimenawasreminiscence": "追憶のしめ�?",
    "silkenmoonsserenade": "月を紡ぐ夜�?��?",
    "nightoftheskysunveiling": "天穹の顕現せし�?",
    "unfinishedreverie": "遂げられなかった想�?",
    "longnightsoath": "長き夜�?�誓い",
    "blizzardstrayer": "氷風を彷徨�?�?士",
    "marechausseehunter": "ファントムハンター",
    "theexile": "亡命�?",
    "crimsonwitchofflames": "�?え盛る炎の魔女",
    "archaicpetra": "�?�?の磐岩",
    "echoesofanoffering": "来�?の余響",
    "flowerofparadiselost": "楽園�?�絶花",
    "lavawalker": "烈火を渡る賢�?"
};

function toJPCharacter(key) {
    if (!key) return key;
    return CHAR_TO_JP[key] || key;
}

function toJPWeapon(key) {
    if (!key) return key;
    return WEAPON_TO_JP[key] || key;
}

function toJPArtifact(key) {
    if (!key) return key;
    return ARTIFACT_TO_JP[key] || key;
}

// -----------------------------------------------------------------------------

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
        // Cap visual height to a sensible absolute maximum to avoid runaway sizes
        const viewportH = (typeof window !== 'undefined' && window.innerHeight) ? window.innerHeight : 800;
        const absoluteMax = Math.max(800, Math.floor(viewportH * 2.5)); // allow tall charts but bounded
        const cappedHeight = Math.min(heightPx, absoluteMax);
        try { if (parent) parent.style.setProperty('min-height', cappedHeight + 'px', 'important'); } catch (e) {}
        canvas.width = Math.floor(visualWidth * dpr);
        canvas.height = Math.floor(cappedHeight * dpr);
        // Store the intended visual height (capped) on the canvas element to avoid measuring
        // live layout which can cause circular sizing increases.
        try { canvas.dataset.visualHeight = String(cappedHeight); } catch (e) {}
        try { ensureContainerHeight(ctx, cappedHeight); } catch(e) {}
        // Try to set an initial top inset immediately to avoid canvas appearing at top of container
        try {
            // setCanvasTopInset is safe to call; it will measure any title/legend already present
            setCanvasTopInset(canvas);
        } catch(e) {}
        // Hide the canvas until insets and parent heights are applied to avoid a brief overlay flicker
        try { canvas.style.visibility = 'hidden'; canvas.dataset.needUnhide = '1'; } catch(e) {}
    } catch (e) { /* ignore sizing errors */ }
}

// Measure title/legend area inside a chart container and set canvas.style.top so
// the canvas doesn't overlap the title/legend. Leaves a small padding gap.
function setCanvasTopInset(canvas) {
    if (!canvas || !canvas.parentElement) return;
    const container = canvas.parentElement;
    // measure any title/legend/table height inside container
    let height = 0;
    const title = container.querySelector('h6');
    const legend = container.querySelector('.chart-legend');
    const table = container.querySelector('.chart-data-table');
    [title, legend, table].forEach(el => {
        if (el && el.getBoundingClientRect) {
            const rect = el.getBoundingClientRect();
            if (rect && rect.height) height += Math.ceil(rect.height);
        }
    });
    // add small padding
    const padding = 12;
    let top = Math.max(8, height + padding);
    // Cap top inset to avoid excessive offset (use a fraction of viewport)
    const vp = (typeof window !== 'undefined' && window.innerHeight) ? window.innerHeight : 800;
    const topCap = Math.max(48, Math.floor(vp * 0.25));
    if (top > topCap) top = topCap;
    try { canvas.style.top = top + 'px'; } catch (e) { /* ignore */ }
    return top;
}

function adjustAllChartInsets() {
    try {
        document.querySelectorAll('.chart-container canvas').forEach(c => {
            const top = setCanvasTopInset(c) || 0;
            try {
                // compute visual canvas height (device-independent pixels)
                const dpr = (typeof window !== 'undefined' && window.devicePixelRatio) ? window.devicePixelRatio : 1;
                // Prefer the visual height stored by setCanvasVisualSize to avoid reading
                // live layout which may reflect transient values.
                const stored = c.dataset && c.dataset.visualHeight ? parseFloat(c.dataset.visualHeight) : NaN;
                let visualCanvasH = 0;
                if (!Number.isNaN(stored) && stored > 0) {
                    visualCanvasH = Math.round(stored);
                } else {
                    const bufHeight = c.height || parseFloat(c.getAttribute('height')) || 0; // actual drawing buffer height
                    visualCanvasH = bufHeight ? Math.round(bufHeight / dpr) : Math.ceil((c.getBoundingClientRect && c.getBoundingClientRect().height) || 0);
                }
                // bottom inset from CSS (fallback to 6px)
                const cs = window.getComputedStyle ? window.getComputedStyle(c) : null;
                const bottomInset = cs ? (parseFloat(cs.bottom) || 6) : 6;
                const required = Math.max(120, Math.ceil(top + visualCanvasH + bottomInset + 2));
                // Prevent exploding required values by capping to an absolute maximum
                const viewportH = (typeof window !== 'undefined' && window.innerHeight) ? window.innerHeight : 800;
                const absoluteMax = Math.max(1000, Math.floor(viewportH * 3));
                const finalRequired = Math.min(required, absoluteMax);
                const parent = c.parentElement;
                if (parent) {
                    try {
                        // Read existing min-height (inline style first, fallback to computed)
                        const existingInline = parent.style && parent.style.minHeight ? parseFloat(parent.style.minHeight) : NaN;
                        const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(parent).minHeight) : NaN;
                        const existing = (!Number.isNaN(existingInline) && existingInline > 0) ? existingInline : (Number.isNaN(computed) ? 0 : computed);
                        // Only increase min-height; cap per-step increase to avoid big jumps
                        const maxStep = Math.max(200, Math.floor(existing * 0.5)); // at most +50% or +200px
                        let newHeight = finalRequired;
                        if (finalRequired > existing && finalRequired - existing > maxStep) {
                            newHeight = existing + maxStep;
                        }
                        if (newHeight > existing) parent.style.setProperty('min-height', Math.ceil(newHeight) + 'px', 'important');
                    } catch (e) { /* ignore */ }
                }
                // ensure ancestor .col also reserves height (only increase)
                let el = parent; let depth = 0;
                while (el && depth < 4) {
                    if (el.classList && el.classList.contains('col')) {
                        try {
                            const existingInline = el.style && el.style.minHeight ? parseFloat(el.style.minHeight) : NaN;
                            const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(el).minHeight) : NaN;
                            const existing = (!Number.isNaN(existingInline) && existingInline > 0) ? existingInline : (Number.isNaN(computed) ? 0 : computed);
                            if (required > existing) el.style.setProperty('min-height', required + 'px', 'important');
                        } catch(e) { /* ignore */ }
                        break;
                    }
                    el = el.parentElement; depth++;
                }
                    // If this canvas was hidden pending inset application, unhide it now
                    try {
                        if (c && c.dataset && c.dataset.needUnhide) {
                            c.style.visibility = 'visible';
                            delete c.dataset.needUnhide;
                        } else if (c && (!c.dataset || !c.dataset.needUnhide)) {
                            // If canvas was hidden via CSS initial state, make it visible now that layout is stable
                            c.style.visibility = 'visible';
                        }
                    } catch(e) { /* ignore */ }
            } catch (e) { /* ignore per-chart errors */ }
        });
    } catch (e) { /* ignore */ }
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
// by duration here �? mean is the canonical numeric statistic returned by the simulator.
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

// If Chart is available, register a small plugin that triggers inset adjustment after rendering
try {
    if (typeof Chart !== 'undefined' && typeof Chart.register === 'function') {
        Chart.register({
            id: 'gcsim-adjust-inset',
            afterRender: function(chart) {
                try { if (typeof adjustAllChartInsets === 'function') adjustAllChartInsets(); } catch(e) {}
            }
        });
    }
} catch (e) { /* ignore if Chart isn't loaded yet */ }

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
    
    // Screen navigation setup
    setupScreenNavigation();
    
    // Mode switching setup
    setupModeSwitch();
    
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

// Global flag to disable initial Chart.js animations to avoid layout shifts on first render
const DISABLE_INITIAL_CHART_ANIMATION = true;
try {
    if (typeof Chart !== 'undefined' && Chart.defaults && Chart.defaults.plugins) {
        if (DISABLE_INITIAL_CHART_ANIMATION) {
            // Turn off animation for initial draw; keep hover/tooltip animations
            Chart.defaults.animation = false;
        }
    }
} catch(e) { /* ignore if Chart not present */ }

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
        
        // Switch to results screen
        const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
        if (resultsTab) {
            resultsTab.click();
        }
        
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
        message = 'パ�?�スエラー:\n';
        error.parse_errors.forEach(pe => {
            if (pe.line) {
                message += `�? ${pe.line}: ${pe.message}\n`;
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
    // Make sure the results screen is visible
    const resultsScreen = document.getElementById('screen-results');
    if (resultsScreen && !resultsScreen.classList.contains('active')) {
        // Switch to results screen
        const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
        if (resultsTab) {
            resultsTab.click();
        }
    }
    
    // Keep results hidden until layout (charts/insets) is applied to avoid brief overlap
    // We'll make it visible after charts are created and insets reserved.
    
    // Display statistics
    displayStatistics(result);
    
    // Display character information
    displayCharacters(result);
    
    // Display target information
    displayTargetInfo(result);
    
    // Display charts (this will set canvas sizes and schedule insets)
    displayCharts(result);

    // After charts are created and insets applied, make the results visible.
    // Use a short timeout to allow Chart.js to run initial layout and our inset adjustments.
    setTimeout(() => {
        try { if (typeof adjustAllChartInsets === 'function') adjustAllChartInsets(); } catch(e) {}
        try { resultsContainer.style.display = 'block'; resultsContainer.classList.add('visible'); } catch(e) {}
        try { resultsContainer.scrollIntoView({ behavior: 'smooth' }); } catch(e) {}
        debugLog('[WebUI] Results displayed successfully (post-layout)');
    }, 80);
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
        container.innerHTML = '<p>キャラクター�?報がありません</p>';
        return;
    }
    
    result.character_details.forEach((char, idx) => {
        console.log(`[WebUI] Character ${idx} keys:`, Object.keys(char));
        console.log(`[WebUI] Character ${idx} data:`, JSON.stringify(char, null, 2));

        const charDiv = document.createElement('div');
        charDiv.className = 'char-card';

        const rawName = char.name || 'Unknown';
        const name = toJPCharacter(rawName);
        const level = char.level || 1;
        const maxLevel = char.max_level || 90;
        const constellation = char.cons || 0;
        const weapon = char.weapon?.name || 'Unknown';
        const weaponJP = toJPWeapon(weapon);
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
                `<span class="chip">${toJPArtifact(set)} (${count})<div class="small-en">${set}</div></span>`
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
                { name: '攻�?�?', value: finalATK, format: (v) => Math.round(v) },
                { name: '防御�?', value: finalDEF, format: (v) => Math.round(v) },
                { name: '�?�?熟知', value: finalEM, format: (v) => Math.round(v) },
                { name: '会�?�?', value: finalCR, format: (v) => (v * 100).toFixed(1) + '%' },
                { name: '会�?ダメージ', value: finalCD, format: (v) => (v * 100).toFixed(1) + '%' },
                { name: '�?�?チャージ効�?', value: finalER, format: (v) => (v * 100).toFixed(1) + '%' },
            ];
            
            console.log(`[WebUI] ${name} final stats: HP=${Math.round(finalHP)}, ATK=${Math.round(finalATK)}, DEF=${Math.round(finalDEF)}, EM=${Math.round(finalEM)}, CR=${(finalCR*100).toFixed(1)}%, CD=${(finalCD*100).toFixed(1)}%, ER=${(finalER*100).toFixed(1)}%`);
            
            statsHTML = '<div class="char-stats-list"><div style="margin: 6px 0; font-size: 0.85rem;"><strong>ス�?ータス詳細:</strong></div>';
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
            <div class="char-name">${name} <span class="char-en">${rawName}</span></div>
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
                    <span class="info-value">${weaponJP} Lv.${weaponLevel}/${weaponMaxLevel} (R${weaponRefine})<div class="small-en">${weapon}</div></span>
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
            targetsDiv.innerHTML = `<div class="card-content"><span class="card-title">ターゲ�?ト情報</span>${targetsBlockHtml}</div>`;
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
        container.innerHTML = '<p>ターゲ�?ト情報がありません</p>';
        return;
    }

    // Render as requested: plain label 'ターゲ�?ト情報:' and for each target a compact block
    const header = document.createElement('div');
    header.className = 'card';
    header.style.padding = '8px';
    header.style.marginBottom = '8px';
    header.innerHTML = `<div style="font-weight:700;">ターゲ�?ト情報:</div>`;
    container.appendChild(header);

    // helper to remove ~~strike~~ tokens and their contents
    function stripStrikeTokens(s) {
        if (!s) return s;
        // remove all occurrences of ~~...~~ (non-greedy)
        return s.replace(/~~.*?~~/g, '').trim();
    }

    result.target_details.forEach((target, idx) => {
        const targetDiv = document.createElement('div');
        targetDiv.className = 'char-card';
        targetDiv.style.display = 'block';
        targetDiv.style.padding = '8px';
        targetDiv.style.marginBottom = '8px';

        const name = stripStrikeTokens(target.name) || `ターゲ�?�? ${idx + 1}`;
        const level = target.level || 1;
        const hp = target.hp || 0;

        // Build resist lines; element names may contain markdown-like ~~strikethrough~~ tokens; keep as-is
        let resistLines = '';
        if (target.resist && Object.keys(target.resist).length > 0) {
            for (const [element, resist] of Object.entries(target.resist)) {
                const el = stripStrikeTokens(element);
                if (!el) continue; // if the label was entirely struck out, skip it
                resistLines += `<div class="info-row"><span class="info-label">${el}</span><span class="info-value">${(resist * 100).toFixed(1)}%</span></div>`;
            }
        }

        targetDiv.innerHTML = `
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;"><div style="font-weight:600;">${name}</div><div>Lv.${level}</div></div>
            <div class="info-row"><span class="info-label">HP</span><span class="info-value">${formatNumber(hp)}</span></div>
            <div style="margin-top:6px;"><strong>耐性:</strong></div>
            ${resistLines}
        `;

        container.appendChild(targetDiv);
    });
}

// Build HTML block for all targets so it can be embedded under each character card
function buildTargetsHTML(result) {
    if (!result.target_details || result.target_details.length === 0) return '';
    let html = '<div style="margin-top:10px;"><strong>ターゲ�?ト情報:</strong>';
    result.target_details.forEach((target, idx) => {
        // reuse stripStrikeTokens if available, otherwise define a local fallback
        const stripStrikeTokens = (typeof stripStrikeTokens === 'function') ? stripStrikeTokens : function(s) { return s ? s.replace(/~~.*?~~/g,'').trim() : s; };
        const name = stripStrikeTokens(target.name) || `ターゲ�?�? ${idx + 1}`;
        const level = target.level || 1;
        const hp = target.hp || 0;
        let resistHTML = '';
        if (target.resist && Object.keys(target.resist).length > 0) {
            resistHTML = '<div style="margin-top:6px;">';
            for (const [element, resist] of Object.entries(target.resist)) {
                const el = stripStrikeTokens(element);
                if (!el) continue;
                resistHTML += `<div class="info-row"><span class="info-label">${el}</span><span class="info-value">${(resist * 100).toFixed(1)}%</span></div>`;
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
                const rawName = char.name || `キャラ${idx+1}`;
                const name = toJPCharacter(rawName);
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
                    // pass an empty labels array so no 'チ�?��?' label appears on the axis
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
    const charNamesRaw = (result.character_details && Array.isArray(result.character_details)) ? result.character_details.map(c => toJPCharacter(c.name)) : stats.source_dps.map((_,i) => `キャラ${i+1}`);
    const charNames = (stats.__char_order && stats.__char_order.orderedCharNames) ? stats.__char_order.orderedCharNames : charNamesRaw;
        // Collect ability/source keys from source_dps
        const abilitySet = new Set();
        stats.source_dps.forEach(sa => { if (sa && sa.sources) Object.keys(sa.sources).forEach(k => abilitySet.add(k)); });
        const abilities = Array.from(abilitySet);

        if (abilities.length > 0) {
            // Create a matrix matching sorted charNames order. source_dps is indexed by original character index,
            // so we need to map canonical ordering indices back to original indices in source_dps.
            const originalCharNames = (result.character_details && Array.isArray(result.character_details)) ? result.character_details.map(c => toJPCharacter(c.name)) : stats.source_dps.map((_,i) => `キャラ${i+1}`);
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
            try { showEmptyChartPlaceholder(ctx2.canvas.parentElement, 'ソース別DPS の�?ータがありません'); } catch(e) {}
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
            try { showEmptyChartPlaceholder(ctx5.canvas.parentElement, '反応回数の�?ータがありません'); } catch(e) {}
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
            try { showEmptyChartPlaceholder(ctx5.canvas.parentElement, '反応回数の�?ータがありません'); } catch(e) {}
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
    const reactionsDesired = Math.max(200, reactions.length * 30);
    ensureContainerHeight(ctx5, reactionsDesired);
    try { setCanvasVisualSize(ctx5, reactionsDesired); } catch(e) {}
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
    try { adjustAllChartInsets(); } catch(e) {}
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
            try { showEmptyChartPlaceholder(ctx6.canvas.parentElement, '付着時間の�?ータがありません'); } catch(e) {}
            return;
        }

        // Each entry in target_aura_uptime represents a target: map of element->value (0..10000)
        const targetLabels = [];
        const elementSet = new Set();
        const perTarget = []; // array of maps element->value

        stats.target_aura_uptime.forEach((targetAura, tidx) => {
            const label = `ターゲ�?�?${tidx+1}`;
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
            try { showEmptyChartPlaceholder(ctx6.canvas.parentElement, '付着時間の�?ータがありません'); } catch(e) {}
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

    const auraDesired = Math.max(200, targetLabels.length * 50);
    ensureContainerHeight(ctx6, auraDesired);
    try { setCanvasVisualSize(ctx6, auraDesired); } catch(e) {}
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
        try { adjustAllChartInsets(); } catch(e) {}
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
        div.textContent = text || '�?ータがありません';
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
    
    let html = '<div class="chart-legend-title">凡�?</div>';
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
        'atk': '攻�?�?',
        'atk%': '攻�?�?%',
        'def': '防御�?',
        'def%': '防御�?%',
        'em': '�?�?熟知',
        'er': '�?�?チャージ効�?',
        'cr': '会�?�?',
        'cd': '会�?ダメージ',
        'heal': '治療効�?',
        'pyro%': '炎�??�?ダメージ',
        'hydro%': '水�?�?ダメージ',
        'cryo%': '氷�?�?ダメージ',
        'electro%': '雷�?�?ダメージ',
        'anemo%': '風�?�?ダメージ',
        'geo%': '岩�?�?ダメージ',
        'dendro%': '草�??�?ダメージ',
        'phys%': '物�?ダメージ'
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

    // After all charts are created, adjust canvas top insets so titles/legends are not overlapped
    try { 
        adjustAllChartInsets();
        // Chart.js may resize/update asynchronously; retry a few times to be safe
        const retries = [120, 300, 600];
        retries.forEach((delay, idx) => {
            setTimeout(() => { try { adjustAllChartInsets(); } catch(e){} }, delay);
        });
    } catch (e) { /* ignore */ }
}

// Recompute insets on window resize (debounced)
let _resizeTimer = null;
window.addEventListener('resize', () => {
    if (_resizeTimer) clearTimeout(_resizeTimer);
    _resizeTimer = setTimeout(() => { adjustAllChartInsets(); }, 150);
});

// Make functions available globally for onclick handlers
window.runSimulation = runSimulation;
window.switchTab = switchTab;
window.runOptimizerSimulation = runOptimizerSimulation;
