(() => {
  // src/app.js
  var charts = {};
  var cmEditor = null;
  var cmEditorOriginal = null;
  var cmEditorOptimized = null;
  function setupScreenNavigation() {
    const tabsContainer = document.querySelector(".navbar-tabs");
    const screens = document.querySelectorAll(".screen");
    if (!tabsContainer)
      return;
    tabsContainer.addEventListener("click", function(e) {
      const btn = e.target.closest(".navbar-tab");
      if (!btn)
        return;
      if (btn.disabled)
        return;
      const screenId = btn.getAttribute("data-screen");
      tabsContainer.querySelectorAll(".navbar-tab").forEach((b) => b.classList.remove("active"));
      btn.classList.add("active");
      screens.forEach((screen) => screen.classList.remove("active"));
      const targetScreen = document.getElementById("screen-" + screenId);
      if (targetScreen) {
        targetScreen.classList.add("active");
        if (screenId === "results" && typeof window.onResultsShown === "function") {
          try {
            window.onResultsShown();
          } catch (e2) {
            console.warn("onResultsShown failed", e2);
          }
        }
      }
    });
  }
  function setupModeSwitch() {
    const modeButtons = document.querySelectorAll(".mode-btn");
    const modes = document.querySelectorAll(".config-mode");
    modeButtons.forEach((button) => {
      button.addEventListener("click", function() {
        const modeId = this.getAttribute("data-mode");
        modeButtons.forEach((btn) => btn.classList.remove("active"));
        this.classList.add("active");
        modes.forEach((mode) => mode.classList.remove("active"));
        const targetMode = document.getElementById("mode-" + modeId);
        if (targetMode) {
          targetMode.classList.add("active");
        }
      });
    });
  }
  async function runOptimizerSimulation() {
    debugLog("[WebUI] Starting optimizer simulation...");
    const originalTextarea = document.getElementById("config-editor-original");
    const optimizedTextarea = document.getElementById("config-editor-optimized");
    const errorMsg = document.getElementById("error-message-optimizer");
    const loading = document.getElementById("loading-optimizer");
    const runButton = document.querySelector("#mode-optimizer .btn-run");
    let config = "";
    if (optimizedTextarea.value.trim()) {
      config = typeof cmEditorOptimized !== "undefined" && cmEditorOptimized ? cmEditorOptimized.getValue() : optimizedTextarea.value;
    } else if (originalTextarea.value.trim()) {
      config = typeof cmEditorOriginal !== "undefined" && cmEditorOriginal ? cmEditorOriginal.getValue() : originalTextarea.value;
    }
    if (!config.trim()) {
      errorMsg.textContent = "\u30A8\u30E9\u30FC: \u30B3\u30F3\u30D5\u30A3\u30B0\u3092\u5165\u529B\u3057\u3066\u304F\u3060\u3055\u3044";
      errorMsg.style.display = "block";
      return;
    }
    debugLog("[WebUI] Config length:", config.length);
    errorMsg.style.display = "none";
    loading.style.display = "block";
    runButton.disabled = true;
    try {
      debugLog("[WebUI] Sending request to /api/optimize");
      const response = await fetch("/api/optimize", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ config })
      });
      debugLog("[WebUI] Response status:", response.status);
      loading.style.display = "none";
      runButton.disabled = false;
      if (!response.ok) {
        const error = await response.json();
        console.error("[WebUI] Error response:", error);
        errorMsg.textContent = "\u30A8\u30E9\u30FC: " + (error.message || error.error || "Optimizer\u5B9F\u884C\u306B\u5931\u6557\u3057\u307E\u3057\u305F");
        errorMsg.style.display = "block";
        return;
      }
      const result = await response.json();
      debugLog("[WebUI] Optimizer result:", result);
      if (result.optimized_config) {
        if (cmEditorOptimized) {
          cmEditorOptimized.setValue(result.optimized_config);
        } else {
          optimizedTextarea.value = result.optimized_config;
        }
      }
      if (result.statistics) {
        const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
        if (resultsTab) {
          resultsTab.click();
        }
        displayResults(result);
      }
    } catch (err) {
      console.error("[WebUI] Exception:", err);
      loading.style.display = "none";
      runButton.disabled = false;
      errorMsg.textContent = "\u30A8\u30E9\u30FC: " + err.message;
      errorMsg.style.display = "block";
    }
  }
  function hexToRgba(hex, alpha) {
    if (!hex)
      return `rgba(0,0,0,${alpha})`;
    const h = hex.replace("#", "");
    const bigint = parseInt(h, 16);
    const r = bigint >> 16 & 255;
    const g = bigint >> 8 & 255;
    const b = bigint & 255;
    return `rgba(${r}, ${g}, ${b}, ${alpha})`;
  }
  var CHAR_PALETTE = ["#4e79a7", "#59a14f", "#f28e2b", "#e15759", "#76b7b2", "#edc949", "#af7aa1", "#ff9da7"];
  function getCharColor(i) {
    return CHAR_PALETTE[i % CHAR_PALETTE.length];
  }
  var DEBUG = false;
  function debugLog(...args) {
    if (DEBUG && console && console.log)
      console.log(...args);
  }
  function hashCode(str) {
    let h = 0;
    if (!str)
      return h;
    for (let i = 0; i < str.length; i++) {
      h = (h << 5) - h + str.charCodeAt(i);
      h |= 0;
    }
    return h;
  }
  function ctxRawValue(context) {
    if (context === void 0 || context === null)
      return 0;
    if (typeof context.raw === "number")
      return context.raw;
    if (context.parsed)
      return context.parsed.x || context.parsed || 0;
    return context.raw || 0;
  }
  var CHAR_TO_JP = typeof window !== "undefined" && window.CHAR_TO_JP ? window.CHAR_TO_JP : {};
  var WEAPON_TO_JP = typeof window !== "undefined" && window.WEAPON_TO_JP ? window.WEAPON_TO_JP : {};
  var ARTIFACT_TO_JP = typeof window !== "undefined" && window.ARTIFACT_TO_JP ? window.ARTIFACT_TO_JP : {};
  function toJPCharacter(key) {
    if (!key)
      return key;
    return CHAR_TO_JP[key] || key;
  }
  function toJPWeapon(key) {
    if (!key)
      return key;
    return WEAPON_TO_JP[key] || key;
  }
  function toJPArtifact(key) {
    if (!key)
      return key;
    return ARTIFACT_TO_JP[key] || key;
  }
  function setCanvasVisualSize(ctx, desiredHeightPx, minWidth = 300) {
    try {
      const canvas = ctx && ctx.canvas ? ctx.canvas : null;
      if (!canvas)
        return;
      const parent = canvas.parentElement;
      const dpr = typeof window !== "undefined" && window.devicePixelRatio ? window.devicePixelRatio : 1;
      try {
        canvas.style.width = "100%";
      } catch (e) {
      }
      const rect = parent && parent.getBoundingClientRect ? parent.getBoundingClientRect() : canvas.getBoundingClientRect();
      const visualWidth = rect && rect.width ? Math.max(minWidth, Math.floor(rect.width)) : Math.max(minWidth, Math.floor(canvas.offsetWidth || 600));
      const heightPx = Math.max(120, Math.floor(desiredHeightPx || 140));
      const viewportH = typeof window !== "undefined" && window.innerHeight ? window.innerHeight : 800;
      const absoluteMax = Math.max(800, Math.floor(viewportH * 2.5));
      const cappedHeight = Math.min(heightPx, absoluteMax);
      try {
        if (parent)
          parent.style.setProperty("min-height", cappedHeight + "px", "important");
      } catch (e) {
      }
      canvas.width = Math.floor(visualWidth * dpr);
      canvas.height = Math.floor(cappedHeight * dpr);
      try {
        canvas.dataset.visualHeight = String(cappedHeight);
      } catch (e) {
      }
      try {
        ensureContainerHeight(ctx, cappedHeight);
      } catch (e) {
      }
      try {
        setCanvasTopInset(canvas);
      } catch (e) {
      }
      try {
        canvas.style.visibility = "hidden";
        canvas.dataset.needUnhide = "1";
      } catch (e) {
      }
    } catch (e) {
    }
  }
  function setCanvasTopInset(canvas) {
    if (!canvas || !canvas.parentElement)
      return;
    const container = canvas.parentElement;
    let height = 0;
    const title = container.querySelector("h6");
    const legend = container.querySelector(".chart-legend");
    const table = container.querySelector(".chart-data-table");
    [title, legend, table].forEach((el) => {
      if (el && el.getBoundingClientRect) {
        const rect = el.getBoundingClientRect();
        if (rect && rect.height)
          height += Math.ceil(rect.height);
      }
    });
    const padding = 12;
    let top = Math.max(8, height + padding);
    const vp = typeof window !== "undefined" && window.innerHeight ? window.innerHeight : 800;
    const topCap = Math.max(48, Math.floor(vp * 0.25));
    if (top > topCap)
      top = topCap;
    try {
      canvas.style.top = top + "px";
    } catch (e) {
    }
    return top;
  }
  function resetChartContainerHeights() {
    try {
      document.querySelectorAll(".chart-container, .chart-container-compact").forEach((el) => {
        try {
          el.style.removeProperty("min-height");
        } catch (e) {
        }
      });
      document.querySelectorAll(".chart-container canvas, .chart-container-compact canvas").forEach((c) => {
        try {
          if (c && c.dataset) {
            delete c.dataset.visualHeight;
            delete c.dataset.needUnhide;
          }
          try {
            c.style.removeProperty("top");
          } catch (e) {
          }
          try {
            c.style.removeProperty("width");
          } catch (e) {
          }
          try {
            c.style.removeProperty("height");
          } catch (e) {
          }
          try {
            c.style.visibility = "";
          } catch (e) {
          }
        } catch (e) {
        }
      });
    } catch (e) {
    }
  }
  function adjustAllChartInsets() {
    try {
      document.querySelectorAll(".chart-container canvas").forEach((c) => {
        const top = setCanvasTopInset(c) || 0;
        try {
          const dpr = typeof window !== "undefined" && window.devicePixelRatio ? window.devicePixelRatio : 1;
          const stored = c.dataset && c.dataset.visualHeight ? parseFloat(c.dataset.visualHeight) : NaN;
          let visualCanvasH = 0;
          if (!Number.isNaN(stored) && stored > 0) {
            visualCanvasH = Math.round(stored);
          } else {
            const bufHeight = c.height || parseFloat(c.getAttribute("height")) || 0;
            visualCanvasH = bufHeight ? Math.round(bufHeight / dpr) : Math.ceil(c.getBoundingClientRect && c.getBoundingClientRect().height || 0);
          }
          const cs = window.getComputedStyle ? window.getComputedStyle(c) : null;
          const bottomInset = cs ? parseFloat(cs.bottom) || 6 : 6;
          const required = Math.max(120, Math.ceil(top + visualCanvasH + bottomInset + 2));
          const viewportH = typeof window !== "undefined" && window.innerHeight ? window.innerHeight : 800;
          const absoluteMax = Math.max(1e3, Math.floor(viewportH * 3));
          const finalRequired = Math.min(required, absoluteMax);
          const parent = c.parentElement;
          if (parent) {
            try {
              const existingInline = parent.style && parent.style.minHeight ? parseFloat(parent.style.minHeight) : NaN;
              const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(parent).minHeight) : NaN;
              const existing = !Number.isNaN(existingInline) && existingInline > 0 ? existingInline : Number.isNaN(computed) ? 0 : computed;
              const maxStep = Math.max(200, Math.floor(existing * 0.5));
              let newHeight = finalRequired;
              if (finalRequired > existing && finalRequired - existing > maxStep) {
                newHeight = existing + maxStep;
              }
              if (newHeight > existing)
                parent.style.setProperty("min-height", Math.ceil(newHeight) + "px", "important");
            } catch (e) {
            }
          }
          let el = parent;
          let depth = 0;
          while (el && depth < 4) {
            if (el.classList && el.classList.contains("col")) {
              try {
                const existingInline = el.style && el.style.minHeight ? parseFloat(el.style.minHeight) : NaN;
                const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(el).minHeight) : NaN;
                const existing = !Number.isNaN(existingInline) && existingInline > 0 ? existingInline : Number.isNaN(computed) ? 0 : computed;
                if (required > existing)
                  el.style.setProperty("min-height", required + "px", "important");
              } catch (e) {
              }
              break;
            }
            el = el.parentElement;
            depth++;
          }
          try {
            if (c && c.dataset && c.dataset.needUnhide) {
              c.style.visibility = "visible";
              delete c.dataset.needUnhide;
            } else if (c && (!c.dataset || !c.dataset.needUnhide)) {
              c.style.visibility = "visible";
            }
          } catch (e) {
          }
        } catch (e) {
        }
      });
    } catch (e) {
    }
  }
  function formatValue(v, decimals = 2, suffix = "") {
    const n = Number(v) || 0;
    const isInt = Math.abs(n - Math.round(n)) < 1e-9;
    const body = isInt ? Math.round(n).toString() : n.toFixed(decimals);
    return body + (suffix || "");
  }
  function extractDescriptiveMean(v) {
    if (v === null || v === void 0)
      return null;
    if (typeof v === "object" && typeof v.mean === "number")
      return v.mean;
    return null;
  }
  function extractNumber(v) {
    if (v === null || v === void 0)
      return 0;
    if (typeof v === "number")
      return Number.isFinite(v) ? v : 0;
    if (typeof v === "object") {
      if (typeof v.mean === "number" && Number.isFinite(v.mean))
        return v.mean;
      if (typeof v.value === "number" && Number.isFinite(v.value))
        return v.value;
      let s = 0;
      let found = false;
      for (const vv of Object.values(v)) {
        if (typeof vv === "number" && Number.isFinite(vv)) {
          s += vv;
          found = true;
        } else if (typeof vv === "object") {
          const nested = extractNumber(vv);
          if (nested) {
            s += nested;
            found = true;
          }
        }
      }
      return found ? s : 0;
    }
    return 0;
  }
  try {
    if (typeof Chart !== "undefined" && Chart.defaults && Chart.defaults.font) {
      Chart.defaults.font.family = "'UDEV Gothic', 'Segoe UI', Arial, sans-serif";
    }
  } catch (e) {
  }
  try {
    if (typeof Chart !== "undefined" && typeof Chart.register === "function") {
      Chart.register({
        id: "gcsim-adjust-inset",
        afterRender: function(chart) {
          try {
            if (typeof adjustAllChartInsets === "function")
              adjustAllChartInsets();
          } catch (e) {
          }
        }
      });
    }
  } catch (e) {
  }
  try {
    if (typeof CodeMirror !== "undefined" && !CodeMirror.modes["gcsl"]) {
      CodeMirror.defineMode("gcsl", function(config, parserConfig) {
        const keywords = /* @__PURE__ */ new Set(["char", "add", "set", "stats", "target", "energy", "active", "options", "if", "else", "for", "while", "return", "break", "continue", "let", "fn", "skill", "burst", "attack", "dash", "charge"]);
        return {
          token: function(stream, state) {
            if (stream.match("//") || stream.match("/*")) {
              if (stream.match("//")) {
                stream.skipToEnd();
                return "comment";
              }
              while (!stream.eol()) {
                if (stream.match("*/"))
                  break;
                stream.next();
              }
              return "comment";
            }
            if (stream.match(/^(?:"(?:[^\\"]|\\.)*"|\'(?:[^\\']|\\.)*\')/)) {
              return "string";
            }
            if (stream.match(/^\d+(?:\.\d+)?/)) {
              return "number";
            }
            if (stream.match(/^[A-Za-z_][A-Za-z0-9_]*/)) {
              const cur = stream.current();
              if (keywords.has(cur))
                return "keyword";
              return "variable";
            }
            stream.next();
            return null;
          }
        };
      });
    }
  } catch (e) {
    console.warn("CodeMirror GCSL mode registration failed", e);
  }
  document.addEventListener("DOMContentLoaded", function() {
    debugLog("[WebUI] Initializing...");
    setupScreenNavigation();
    setupModeSwitch();
    const textarea = document.getElementById("config-editor");
    try {
      cmEditor = CodeMirror.fromTextArea(textarea, {
        mode: "gcsl",
        lineNumbers: true,
        lineWrapping: true,
        theme: document.documentElement.getAttribute("data-theme") === "dark" ? "material" : "default",
        tabSize: 2,
        indentWithTabs: false,
        autofocus: true
      });
      cmEditor.addKeyMap({ "Ctrl-Enter": runSimulation, "Cmd-Enter": runSimulation });
      cmEditor.on("change", () => {
        textarea.value = cmEditor.getValue();
      });
      const cmWrapper = cmEditor.getWrapperElement();
      if (cmWrapper) {
        cmWrapper.style.height = "720px";
        cmWrapper.style.fontSize = "13px";
      }
      const scroller = cmWrapper.querySelector(".CodeMirror-scroll");
      if (scroller)
        scroller.style.height = "720px";
      cmEditor.refresh();
    } catch (e) {
      console.warn("CodeMirror init failed, falling back to textarea", e);
      cmEditor = null;
    }
    const savedTheme = localStorage.getItem("gcsim_theme");
    if (savedTheme) {
      document.documentElement.setAttribute("data-theme", savedTheme);
    }
    const themeBtn = document.getElementById("theme-toggle");
    if (themeBtn) {
      themeBtn.addEventListener("click", () => {
        const cur = document.documentElement.getAttribute("data-theme");
        const next = cur === "dark" ? "" : "dark";
        if (next) {
          document.documentElement.setAttribute("data-theme", next);
          localStorage.setItem("gcsim_theme", next);
        } else {
          document.documentElement.removeAttribute("data-theme");
          localStorage.removeItem("gcsim_theme");
        }
      });
    }
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
    if (cmEditor)
      cmEditor.setValue(defaultConfig);
    debugLog("[WebUI] Default config loaded");
    if (!cmEditor) {
      textarea.addEventListener("keydown", function(e) {
        if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
          e.preventDefault();
          runSimulation();
        }
        if (e.key === "Tab") {
          e.preventDefault();
          const start = this.selectionStart;
          const end = this.selectionEnd;
          this.value = this.value.substring(0, start) + "  " + this.value.substring(end);
          this.selectionStart = this.selectionEnd = start + 2;
        }
      });
    }
    const highlightEl = document.getElementById("config-highlight");
    function escapeHtml(s) {
      return s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    }
    function highlightGcsl(text) {
      if (!text)
        return "";
      const rules = [
        { regex: /(?:\/\*[\s\S]*?\*\/|\/\/.*?(?:\n|$))/y, cls: "gcsl-comment" },
        { regex: /(?:\"(?:[^\\\"]|\\.)*\"|\'(?:[^\\\']|\\.)*\')/y, cls: "gcsl-string" },
        { regex: /\b\d+(?:\.\d+)?\b/y, cls: "gcsl-number" },
        { regex: /\b(?:char|add|set|stats|target|energy|active|options|if|else|for|while|return|break|continue|let|fn|skill|burst|attack|dash|charge)\b/g, cls: "gcsl-keyword" },
        { regex: /\b([A-Za-z_][A-Za-z0-9_]*)/y, cls: "gcsl-fn" },
        { regex: /[+\-*/=<>!:]+/y, cls: "gcsl-operator" },
        { regex: /\s+/y, cls: null },
        { regex: /./y, cls: null }
      ];
      let pos = 0;
      let out = "";
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
          out += escapeHtml(src[pos]);
          pos++;
        }
      }
      return out;
    }
    function updateHighlight() {
      const val = typeof cmEditor !== "undefined" && cmEditor ? cmEditor.getValue() : textarea.value;
      if (highlightEl) {
        const codeEl = highlightEl.querySelector("code");
        if (window.Prism && typeof window.Prism.highlightElement === "function" && codeEl) {
          codeEl.textContent = val + "\n";
          try {
            window.Prism.highlightElement(codeEl);
            if (!codeEl.querySelector(".token")) {
              codeEl.innerHTML = highlightGcsl(val) + "\n";
            }
          } catch (e) {
            codeEl.innerHTML = highlightGcsl(val) + "\n";
          }
        } else {
          const code = highlightEl.querySelector("code");
          if (code) {
            code.innerHTML = highlightGcsl(val) + "\n";
          } else {
            highlightEl.innerHTML = highlightGcsl(val) + "\n";
          }
        }
      }
    }
    if (!cmEditor) {
      textarea.addEventListener("scroll", () => {
        if (highlightEl)
          highlightEl.scrollTop = textarea.scrollTop;
      });
      textarea.addEventListener("input", updateHighlight);
    } else {
      cmEditor.on("change", updateHighlight);
      cmEditor.on("scroll", () => {
        if (highlightEl)
          highlightEl.scrollTop = cmEditor.getScrollInfo().top;
      });
    }
    updateHighlight();
  });
  var DISABLE_INITIAL_CHART_ANIMATION = true;
  try {
    if (typeof Chart !== "undefined" && Chart.defaults && Chart.defaults.plugins) {
      if (DISABLE_INITIAL_CHART_ANIMATION) {
        Chart.defaults.animation = false;
      }
    }
  } catch (e) {
  }
  function switchTab(tabId) {
    document.querySelectorAll(".tab-content").forEach((content) => {
      content.classList.remove("active");
    });
    document.querySelectorAll(".tab-btn").forEach((btn) => {
      btn.classList.remove("active");
    });
    document.getElementById(tabId).classList.add("active");
    event.target.classList.add("active");
  }
  function clearErrorHighlights() {
  }
  async function runSimulation() {
    debugLog("[WebUI] Starting simulation...");
    const textarea = document.getElementById("config-editor");
    const config = typeof cmEditor !== "undefined" && cmEditor ? cmEditor.getValue() : textarea.value;
    const errorMsg = document.getElementById("error-message");
    const loading = document.getElementById("loading");
    const resultsContainer2 = document.getElementById("results-container");
    const runButton = document.querySelector(".btn-run");
    debugLog("[WebUI] Config length:", config.length);
    errorMsg.style.display = "none";
    resultsContainer2.style.display = "none";
    loading.style.display = "block";
    runButton.disabled = true;
    clearErrorHighlights();
    try {
      debugLog("[WebUI] Sending request to /api/simulate");
      const response = await fetch("/api/simulate", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ config })
      });
      debugLog("[WebUI] Response status:", response.status);
      loading.style.display = "none";
      runButton.disabled = false;
      if (!response.ok) {
        const error = await response.json();
        console.error("[WebUI] Error response:", error);
        handleError(error);
        return;
      }
      const result = await response.json();
      debugLog("[WebUI] Simulation result:", result);
      const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
      if (resultsTab) {
        resultsTab.click();
      }
      displayResults(result);
    } catch (err) {
      console.error("[WebUI] Exception:", err);
      loading.style.display = "none";
      runButton.disabled = false;
      errorMsg.textContent = "\u30A8\u30E9\u30FC: " + err.message;
      errorMsg.style.display = "block";
    }
  }
  function handleError(error) {
    debugLog("[WebUI] Handling error:", error);
    const errorMsg = document.getElementById("error-message");
    let message = error.message || error.error || "\u30B7\u30DF\u30E5\u30EC\u30FC\u30B7\u30E7\u30F3\u306B\u5931\u6557\u3057\u307E\u3057\u305F";
    if (error.parse_errors && error.parse_errors.length > 0) {
      message = "\u30D1\u30FC\u30B9\u30A8\u30E9\u30FC:\n";
      error.parse_errors.forEach((pe) => {
        if (pe.line) {
          message += `\u884C ${pe.line}: ${pe.message}
`;
        } else {
          message += `${pe.message}
`;
        }
      });
    }
    errorMsg.textContent = message;
    errorMsg.style.display = "block";
  }
  function displayResults(result) {
    debugLog("[WebUI] Displaying results...");
    const resultsContainer2 = document.getElementById("results-container");
    const resultsScreen = document.getElementById("screen-results");
    if (resultsScreen && !resultsScreen.classList.contains("active")) {
      const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
      if (resultsTab) {
        resultsTab.click();
      }
    }
    displayStatistics(result);
    displayCharacters(result);
    displayTargetInfo(result);
    displayCharts(result);
    setTimeout(() => {
      try {
        if (typeof adjustAllChartInsets === "function")
          adjustAllChartInsets();
      } catch (e) {
      }
      try {
        resultsContainer2.style.display = "block";
        resultsContainer2.classList.add("visible");
      } catch (e) {
      }
      try {
        resultsContainer2.scrollIntoView({ behavior: "smooth" });
      } catch (e) {
      }
      debugLog("[WebUI] Results displayed successfully (post-layout)");
    }, 80);
  }
  function displayStatistics(result) {
    var _a, _b, _c, _d, _e, _f, _g, _h, _i, _j, _k, _l, _m;
    debugLog("[WebUI] Displaying statistics...");
    const stats = result.statistics || {};
    const dps = ((_a = stats.dps) == null ? void 0 : _a.mean) || 0;
    const dpsStd = ((_b = stats.dps) == null ? void 0 : _b.sd) || 0;
    const eps = ((_c = stats.eps) == null ? void 0 : _c.mean) || 0;
    const epsStd = ((_d = stats.eps) == null ? void 0 : _d.sd) || 0;
    const rps = ((_e = stats.rps) == null ? void 0 : _e.mean) || 0;
    const rpsStd = ((_f = stats.rps) == null ? void 0 : _f.sd) || 0;
    const hps = ((_g = stats.hps) == null ? void 0 : _g.mean) || 0;
    const hpsStd = ((_h = stats.hps) == null ? void 0 : _h.sd) || 0;
    const shp = ((_i = stats.shp) == null ? void 0 : _i.mean) || 0;
    const shpStd = ((_j = stats.shp) == null ? void 0 : _j.sd) || 0;
    const duration = ((_k = stats.duration) == null ? void 0 : _k.mean) || ((_l = result.simulator_settings) == null ? void 0 : _l.duration) || 0;
    const durationStd = ((_m = stats.duration) == null ? void 0 : _m.sd) || 0;
    debugLog("[WebUI] Stats:", { dps, eps, rps, hps, shp, duration });
    document.getElementById("stat-dps").innerHTML = formatStatWithStdev(dps, dpsStd);
    document.getElementById("stat-eps").innerHTML = formatStatWithStdev(eps, epsStd);
    document.getElementById("stat-rps").innerHTML = formatStatWithStdev(rps, rpsStd);
    document.getElementById("stat-hps").innerHTML = formatStatWithStdev(hps, hpsStd);
    document.getElementById("stat-shp").innerHTML = formatStatWithStdev(shp, shpStd);
    document.getElementById("stat-dur").innerHTML = formatStatWithStdev(duration, durationStd);
  }
  function displayCharacters(result) {
    debugLog("[WebUI] Displaying characters...");
    debugLog("[WebUI] Full result keys:", Object.keys(result));
    const container = document.getElementById("characters-list");
    container.innerHTML = "";
    const gridDiv = document.createElement("div");
    gridDiv.className = "characters-grid";
    if (!result.character_details || result.character_details.length === 0) {
      container.innerHTML = "<p>\u30AD\u30E3\u30E9\u30AF\u30BF\u30FC\u60C5\u5831\u304C\u3042\u308A\u307E\u305B\u3093</p>";
      return;
    }
    result.character_details.forEach((char, idx) => {
      var _a, _b, _c, _d;
      console.log(`[WebUI] Character ${idx} keys:`, Object.keys(char));
      console.log(`[WebUI] Character ${idx} data:`, JSON.stringify(char, null, 2));
      const charDiv = document.createElement("div");
      charDiv.className = "char-card";
      const rawName = char.name || "Unknown";
      const name = toJPCharacter(rawName);
      const level = char.level || 1;
      const maxLevel = char.max_level || 90;
      const constellation = char.cons || 0;
      const weapon = ((_a = char.weapon) == null ? void 0 : _a.name) || "Unknown";
      const weaponJP = toJPWeapon(weapon);
      const weaponLevel = ((_b = char.weapon) == null ? void 0 : _b.level) || 1;
      const weaponMaxLevel = ((_c = char.weapon) == null ? void 0 : _c.max_level) || 90;
      const weaponRefine = ((_d = char.weapon) == null ? void 0 : _d.refine) || 1;
      const talents = char.talents || {};
      let talentsText = "-";
      if (talents.attack || talents.skill || talents.burst) {
        talentsText = `${talents.attack || 1}/${talents.skill || 1}/${talents.burst || 1}`;
      }
      let firstSetBadge = "";
      let weaponBadgeHTML = "";
      if (char.sets && Object.keys(char.sets).length > 0) {
        const firstSet = Object.entries(char.sets)[0];
        const [set, count] = firstSet;
        firstSetBadge = `<span class="chip">${toJPArtifact(set)} (${count})<div class="small-en">${set}</div></span>`;
      }
      let statsHTML = "";
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
          { name: "HP", value: finalHP, format: (v) => Math.round(v) },
          { name: "\u653B\u6483\u529B", value: finalATK, format: (v) => Math.round(v) },
          { name: "\u9632\u5FA1\u529B", value: finalDEF, format: (v) => Math.round(v) },
          { name: "\u5143\u7D20\u719F\u77E5", value: finalEM, format: (v) => Math.round(v) },
          { name: "\u4F1A\u5FC3\u7387", value: finalCR, format: (v) => (v * 100).toFixed(1) + "%" },
          { name: "\u4F1A\u5FC3\u30C0\u30E1\u30FC\u30B8", value: finalCD, format: (v) => (v * 100).toFixed(1) + "%" },
          { name: "\u5143\u7D20\u30C1\u30E3\u30FC\u30B8\u52B9\u7387", value: finalER, format: (v) => (v * 100).toFixed(1) + "%" }
        ];
        console.log(`[WebUI] ${name} final stats: HP=${Math.round(finalHP)}, ATK=${Math.round(finalATK)}, DEF=${Math.round(finalDEF)}, EM=${Math.round(finalEM)}, CR=${(finalCR * 100).toFixed(1)}%, CD=${(finalCD * 100).toFixed(1)}%, ER=${(finalER * 100).toFixed(1)}%`);
        statsHTML = "";
        statDefs.forEach(({ name: name2, value, format }) => {
          if (value !== void 0 && value !== 0) {
            statsHTML += `<div class="info-row">
                        <span class="info-label">${name2}</span>
                        <span class="info-value">${format(value)}</span>
                    </div>`;
          }
        });
      } else {
        console.log("[WebUI] No snapshot_stats found for character:", name);
      }
      charDiv.innerHTML = `
            <div class="char-header">
                <div class="char-name-line">${name} <span class="char-constellation">C${constellation}</span></div>
                <div class="char-talents">${talentsText}</div>
            </div>
            <div class="char-subheader">
                <div class="char-en-name">${rawName}</div>
                <div class="char-level">Lv. ${level}/${maxLevel}</div>
            </div>
            <div class="char-artifact">
                ${firstSetBadge}
            </div>
            <div class="char-weapon">
                <div class="char-weapon-name">${weaponJP} Lv.${weaponLevel}/${weaponMaxLevel} (R${weaponRefine})</div>
                <div class="char-weapon-en">${weapon}</div>
            </div>
            <div class="char-stats">
                <div class="char-stats-title">\u30B9\u30C6\u30FC\u30BF\u30B9\u8A73\u7D30:</div>
                <div class="char-stats-list">
                    ${statsHTML}
                </div>
            </div>
        `;
      gridDiv.appendChild(charDiv);
    });
    container.appendChild(gridDiv);
    try {
      const targetsBlockHtml = buildTargetsHTML(result);
      if (targetsBlockHtml && targetsBlockHtml.trim().length > 0) {
        const targetsDiv = document.createElement("div");
        targetsDiv.className = "card";
        targetsDiv.style.marginTop = "12px";
        targetsDiv.innerHTML = `<div class="card-content"><span class="card-title">\u30BF\u30FC\u30B2\u30C3\u30C8\u60C5\u5831</span>${targetsBlockHtml}</div>`;
        container.appendChild(targetsDiv);
      }
    } catch (e) {
      console.warn("[WebUI] Could not append targets under characters", e);
    }
  }
  function displayTargetInfo(result) {
    const container = document.getElementById("target-details");
    if (!container)
      return;
    container.innerHTML = "";
    if (!result.target_details || result.target_details.length === 0) {
      container.innerHTML = "<p>\u30BF\u30FC\u30B2\u30C3\u30C8\u60C5\u5831\u304C\u3042\u308A\u307E\u305B\u3093</p>";
      return;
    }
    const header = document.createElement("div");
    header.className = "card";
    header.style.padding = "8px";
    header.style.marginBottom = "8px";
    header.innerHTML = `<div style="font-weight:700;">\u30BF\u30FC\u30B2\u30C3\u30C8\u60C5\u5831:</div>`;
    container.appendChild(header);
    function stripStrikeTokens(s) {
      if (!s)
        return s;
      return s.replace(/~~.*?~~/g, "").trim();
    }
    result.target_details.forEach((target, idx) => {
      const targetDiv = document.createElement("div");
      targetDiv.className = "char-card";
      targetDiv.style.display = "block";
      targetDiv.style.padding = "8px";
      targetDiv.style.marginBottom = "8px";
      const name = stripStrikeTokens(target.name) || `\u30BF\u30FC\u30B2\u30C3\u30C8 ${idx + 1}`;
      const level = target.level || 1;
      const hp = target.hp || 0;
      let resistLines = "";
      if (target.resist && Object.keys(target.resist).length > 0) {
        for (const [element, resist] of Object.entries(target.resist)) {
          const el = stripStrikeTokens(element);
          if (!el)
            continue;
          resistLines += `<div class="info-row"><span class="info-label">${el}</span><span class="info-value">${(resist * 100).toFixed(1)}%</span></div>`;
        }
      }
      targetDiv.innerHTML = `
            <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;"><div style="font-weight:600;">${name}</div><div>Lv.${level}</div></div>
            <div class="info-row"><span class="info-label">HP</span><span class="info-value">${formatNumber(hp)}</span></div>
            <div style="margin-top:6px;"><strong>\u8010\u6027:</strong></div>
            ${resistLines}
        `;
      container.appendChild(targetDiv);
    });
  }
  function buildTargetsHTML(result) {
    if (!result.target_details || result.target_details.length === 0)
      return "";
    let html = '<div style="margin-top:10px;"><strong>\u30BF\u30FC\u30B2\u30C3\u30C8\u60C5\u5831:</strong>';
    // Local helper to strip strikethrough tokens - use different name to avoid TDZ
    const _stripStrikeTokens = function(s) {
      return s ? s.replace(/~~.*?~~/g, "").trim() : s;
    };
    result.target_details.forEach((target, idx) => {
      const name = _stripStrikeTokens(target.name) || `\u30BF\u30FC\u30B2\u30C3\u30C8 ${idx + 1}`;
      const level = target.level || 1;
      const hp = target.hp || 0;
      let resistHTML = "";
      if (target.resist && Object.keys(target.resist).length > 0) {
        resistHTML = '<div style="margin-top:6px;">';
        for (const [element, resist] of Object.entries(target.resist)) {
          const el = _stripStrikeTokens(element);
          if (!el)
            continue;
          resistHTML += `<div class="info-row"><span class="info-label">${el}</span><span class="info-value">${(resist * 100).toFixed(1)}%</span></div>`;
        }
        resistHTML += "</div>";
      }
      html += `<div style="margin-top:8px; padding:8px; border:1px solid var(--muted-border); border-radius:6px; background:var(--card-bg);">
            <div class="info-row"><span class="info-label">${name}</span><span class="info-value">Lv.${level}</span></div>
            <div class="info-row"><span class="info-label">HP</span><span class="info-value">${formatNumber(hp)}</span></div>
            ${resistHTML}
        </div>`;
    });
    html += "</div>";
    return html;
  }
  function displayCharts(result) {
    console.log("[WebUI] Displaying charts...");
    console.log("[WebUI] Result structure:", Object.keys(result));
    console.log("[WebUI] Statistics:", result.statistics);
    
    // Ensure resultsContainer is available at the start (fix for TDZ issue)
    const resultsContainer = document.getElementById('results-container');
    if (!resultsContainer) {
      console.error('[WebUI] results-container not found');
      return;
    }
    
    try {
      resetChartContainerHeights();
    } catch (e) {
    }
    Object.values(charts).forEach((chart) => {
      if (chart && typeof chart.destroy === "function")
        chart.destroy();
    });
    charts = {};
    try {
      resetChartContainerHeights();
    } catch (e) {
    }
    const stats = result.statistics || {};
    try {
      let rawPanel = document.getElementById("raw-stats-panel");
      if (!rawPanel) {
        rawPanel = document.createElement("details");
        rawPanel.id = "raw-stats-panel";
        rawPanel.style.margin = "10px 0";
        const summary = document.createElement("summary");
        summary.textContent = "Raw statistics JSON (debug)";
        rawPanel.appendChild(summary);
        const pre = document.createElement("pre");
        pre.id = "raw-stats-pre";
        pre.style.maxHeight = "300px";
        pre.style.overflow = "auto";
        pre.style.background = "var(--card-bg)";
        pre.style.border = "1px solid var(--muted-border)";
        pre.style.padding = "8px";
        rawPanel.appendChild(pre);
        resultsContainer.insertBefore(rawPanel, resultsContainer.firstChild);
      }
      const preEl = document.getElementById("raw-stats-pre");
      if (preEl)
        preEl.textContent = JSON.stringify(result.statistics || {}, null, 2);
    } catch (e) {
      console.warn("[WebUI] Could not render raw stats panel", e);
    }
    if (result.character_details && result.character_details.length > 0) {
      const canvas = document.getElementById("char-dps-chart");
      if (!canvas) {
        console.error("[WebUI] Canvas element char-dps-chart not found");
      } else {
        console.log("[WebUI] Found canvas element:", canvas);
        const ctx = canvas.getContext("2d");
        console.log("[WebUI] Got 2d context:", ctx);
        const charNames = [];
        const charDpsData = [];
        const charDpsSd = [];
        result.character_details.forEach((char, idx) => {
          var _a, _b, _c, _d;
          const rawName = char.name || `\u30AD\u30E3\u30E9${idx + 1}`;
          const name = toJPCharacter(rawName);
          charNames.push(name);
          let dpsValue = 0;
          let sdValue = 0;
          if (stats.character_dps && Array.isArray(stats.character_dps)) {
            dpsValue = ((_a = stats.character_dps[idx]) == null ? void 0 : _a.mean) || 0;
            sdValue = typeof ((_b = stats.character_dps[idx]) == null ? void 0 : _b.sd) !== "undefined" ? stats.character_dps[idx].sd : 0;
          } else if (stats.character_dps && typeof stats.character_dps === "object") {
            dpsValue = ((_c = stats.character_dps[name]) == null ? void 0 : _c.mean) || 0;
            sdValue = typeof ((_d = stats.character_dps[name]) == null ? void 0 : _d.sd) !== "undefined" ? stats.character_dps[name].sd : 0;
          }
          charDpsData.push(dpsValue);
          charDpsSd.push(sdValue);
        });
        console.log("[WebUI] Character DPS data:", charNames, charDpsData);
        if (charDpsData.length > 0 && charDpsData.some((v) => v > 0)) {
          const order = charDpsData.map((v, i) => ({ idx: i, dps: v }));
          order.sort((a, b) => b.dps - a.dps);
          const orderedCharNames = order.map((o) => charNames[o.idx]);
          const orderedCharDps = order.map((o) => charDpsData[o.idx]);
          const orderedCharSd = order.map((o) => charDpsSd[o.idx]);
          stats.__char_order = { order, orderedCharNames, orderedCharDps, orderedCharSd };
          charts.charDps = createStackedBarChart(ctx, [""], [orderedCharNames, orderedCharDps, orderedCharSd], "\u30AD\u30E3\u30E9\u30AF\u30BF\u30FC\u5225DPS");
        } else {
          console.log("[WebUI] No character DPS data to display");
        }
      }
    }
    const canvas2 = document.getElementById("source-dps-chart");
    if (!canvas2) {
      console.error("[WebUI] Canvas element source-dps-chart not found");
    } else {
      const ctx2 = canvas2.getContext("2d");
      let sourceData = {};
      const durationMean = stats.duration && typeof stats.duration.mean === "number" ? stats.duration.mean : result.simulator_settings && result.simulator_settings.duration || 0;
      if (stats.dps_by_element && Array.isArray(stats.dps_by_element)) {
        stats.dps_by_element.forEach((charData, idx) => {
          var _a, _b;
          const charName = ((_b = (_a = result.character_details) == null ? void 0 : _a[idx]) == null ? void 0 : _b.name) || `\u30AD\u30E3\u30E9${idx + 1}`;
          if (charData && typeof charData === "object") {
            Object.entries(charData).forEach(([element, data2]) => {
              const key = `${charName} (${element})`;
              const mean = extractDescriptiveMean(data2);
              if (typeof mean === "number" && mean > 0)
                sourceData[key] = mean;
            });
          }
        });
      } else if (stats.source_dps && Array.isArray(stats.source_dps)) {
        stats.source_dps.forEach((sa, idx) => {
          var _a, _b;
          const charName = ((_b = (_a = result.character_details) == null ? void 0 : _a[idx]) == null ? void 0 : _b.name) || `\u30AD\u30E3\u30E9${idx + 1}`;
          if (sa && sa.sources) {
            Object.entries(sa.sources).forEach(([source, ds]) => {
              const mean = extractDescriptiveMean(ds);
              const num = mean !== null ? mean : extractNumber(ds);
              if (typeof num === "number" && num > 0)
                sourceData[`${charName}: ${source}`] = num;
            });
          }
        });
      } else {
        if (stats.source_damage_instances)
          console.log("[WebUI] source_damage_instances present but ignored (counts only)");
      }
      console.log("[WebUI] Source DPS data:", sourceData);
      const data = extractChartData(sourceData);
      if (stats.source_dps && Array.isArray(stats.source_dps) && stats.source_dps.length > 0) {
        const charNamesRaw = result.character_details && Array.isArray(result.character_details) ? result.character_details.map((c) => toJPCharacter(c.name)) : stats.source_dps.map((_, i) => `\u30AD\u30E3\u30E9${i + 1}`);
        const charNames = stats.__char_order && stats.__char_order.orderedCharNames ? stats.__char_order.orderedCharNames : charNamesRaw;
        const abilitySet = /* @__PURE__ */ new Set();
        stats.source_dps.forEach((sa) => {
          if (sa && sa.sources)
            Object.keys(sa.sources).forEach((k) => abilitySet.add(k));
        });
        const abilities = Array.from(abilitySet);
        if (abilities.length > 0) {
          const originalCharNames = result.character_details && Array.isArray(result.character_details) ? result.character_details.map((c) => toJPCharacter(c.name)) : stats.source_dps.map((_, i) => `\u30AD\u30E3\u30E9${i + 1}`);
          const canonicalToOriginal = [];
          if (stats.__char_order && stats.__char_order.order) {
            stats.__char_order.order.forEach((o) => canonicalToOriginal.push(o.idx));
          } else {
            for (let i = 0; i < originalCharNames.length; i++)
              canonicalToOriginal.push(i);
          }
          const matrix = abilities.map(() => Array(canonicalToOriginal.length).fill(0));
          const metaMatrix = abilities.map(() => Array(canonicalToOriginal.length).fill(null));
          abilities.forEach((ability, aIdx) => {
            canonicalToOriginal.forEach((origIdx, cCanonicalIdx) => {
              const sa = stats.source_dps[origIdx];
              if (!sa || !sa.sources)
                return;
              const ds = sa.sources[ability];
              if (!ds)
                return;
              const mean = typeof ds.mean === "number" ? ds.mean : 0;
              const sd = typeof ds.sd === "number" ? ds.sd : 0;
              const min = typeof ds.min === "number" ? ds.min : 0;
              const max = typeof ds.max === "number" ? ds.max : 0;
              matrix[aIdx][cCanonicalIdx] = mean;
              metaMatrix[aIdx][cCanonicalIdx] = { mean, sd, min, max };
            });
          });
          charts.sourceDps = createStackedAbilitiesChart(ctx2, charNames, abilities, matrix, "\u30AD\u30E3\u30E9\u30AF\u30BF\u30FC\u5225 \u80FD\u529BDPS", metaMatrix, { barThickness: 24, verticalPadding: 5 });
        } else if (data.labels.length > 0) {
          charts.sourceDps = createBarChart(ctx2, data.labels, data.values, "\u30BD\u30FC\u30B9\u5225DPS");
        } else {
          console.log("[WebUI] No source DPS data to display");
          try {
            showEmptyChartPlaceholder(ctx2.canvas.parentElement, "\u30BD\u30FC\u30B9\u5225DPS \u306E\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093");
          } catch (e) {
          }
        }
      } else if (stats.character_actions && Array.isArray(stats.character_actions) && stats.character_actions.length > 0) {
        console.log("[WebUI] character_actions present but ignored for DPS (contains counts)");
        if (data.labels.length > 0)
          charts.sourceDps = createBarChart(ctx2, data.labels, data.values, "\u30BD\u30FC\u30B9\u5225DPS");
        else
          console.log("[WebUI] No source DPS data to display");
      } else {
        if (data.labels.length > 0)
          charts.sourceDps = createBarChart(ctx2, data.labels, data.values, "\u30BD\u30FC\u30B9\u5225DPS");
        else
          console.log("[WebUI] No source DPS data to display");
      }
    }
    const canvas3 = document.getElementById("damage-dist-chart");
    if (!canvas3) {
      console.error("[WebUI] Canvas element damage-dist-chart not found");
    } else {
      const ctx3 = canvas3.getContext("2d");
      if (stats.damage_buckets) {
        const buckets = stats.damage_buckets;
        const bucketSize = buckets.bucket_size || 30;
        const bucketData = buckets.buckets || [];
        const timeLabels = bucketData.map((_, idx) => {
          const frames = idx * bucketSize;
          const secs = frames / 60;
          return secs >= 1 ? `${secs.toFixed(0)}s` : `${secs.toFixed(2)}s`;
        });
        const damageValues = bucketData.map((bucket) => (bucket == null ? void 0 : bucket.mean) || 0);
        console.log("[WebUI] Damage distribution data:", timeLabels.length, "buckets");
        if (timeLabels.length > 0) {
          charts.damageDist = createLineChart(ctx3, timeLabels, damageValues, "\u30C0\u30E1\u30FC\u30B8", { heightPx: 480 });
        }
      } else {
        console.log("[WebUI] No damage distribution data");
      }
    }
    (function() {
      const canvas5 = document.getElementById("reaction-count-chart");
      if (!canvas5) {
        console.error("[WebUI] Canvas element reaction-count-chart not found");
        return;
      }
      const ctx5 = canvas5.getContext("2d");
      if (!(stats.source_reactions && Array.isArray(stats.source_reactions))) {
        try {
          showEmptyChartPlaceholder(ctx5.canvas.parentElement, "\u53CD\u5FDC\u56DE\u6570\u306E\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093");
        } catch (e) {
        }
        return;
      }
      const reactionsSet = /* @__PURE__ */ new Set();
      const charNames = [];
      const perCharReactions = [];
      stats.source_reactions.forEach((charReactions, idx) => {
        var _a, _b;
        const charName = ((_b = (_a = result.character_details) == null ? void 0 : _a[idx]) == null ? void 0 : _b.name) || `\u30AD\u30E3\u30E9${idx + 1}`;
        charNames.push(charName);
        const map = {};
        if (charReactions && typeof charReactions === "object") {
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
        try {
          showEmptyChartPlaceholder(ctx5.canvas.parentElement, "\u53CD\u5FDC\u56DE\u6570\u306E\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093");
        } catch (e) {
        }
        return;
      }
      const datasets = charNames.map((cn, ci) => {
        const data = reactions.map((r) => perCharReactions[ci][r] || 0);
        return {
          label: cn,
          data,
          backgroundColor: getCharColor(ci),
          stack: "Stack 0"
        };
      });
      const reactionsDesired = Math.max(200, reactions.length * 30);
      ensureContainerHeight(ctx5, reactionsDesired);
      try {
        setCanvasVisualSize(ctx5, reactionsDesired);
      } catch (e) {
      }
      charts.reactions = new Chart(ctx5, {
        type: "bar",
        data: { labels: reactions, datasets },
        options: {
          // Horizontal bars: categories on Y axis
          indexAxis: "y",
          plugins: {
            legend: { position: "top" },
            tooltip: {
              callbacks: {
                label: function(context) {
                  const raw = ctxRawValue(context);
                  return `${context.dataset.label || ""}: ${formatValue(raw, 2)}`;
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
      try {
        scheduleChartResize(charts.reactions, ctx5);
      } catch (e) {
      }
      try {
        adjustAllChartInsets();
      } catch (e) {
      }
    })();
    (function() {
      const canvas6 = document.getElementById("aura-uptime-chart");
      if (!canvas6) {
        console.error("[WebUI] Canvas element aura-uptime-chart not found");
        return;
      }
      const ctx6 = canvas6.getContext("2d");
      if (!(stats.target_aura_uptime && Array.isArray(stats.target_aura_uptime))) {
        try {
          showEmptyChartPlaceholder(ctx6.canvas.parentElement, "\u4ED8\u7740\u6642\u9593\u306E\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093");
        } catch (e) {
        }
        return;
      }
      const targetLabels = [];
      const elementSet = /* @__PURE__ */ new Set();
      const perTarget = [];
      stats.target_aura_uptime.forEach((targetAura, tidx) => {
        const label = `\u30BF\u30FC\u30B2\u30C3\u30C8 ${tidx + 1}`;
        targetLabels.push(label);
        const map = {};
        if (targetAura && typeof targetAura === "object") {
          const inner = targetAura.sources && typeof targetAura.sources === "object" ? targetAura.sources : targetAura;
          Object.entries(inner).forEach(([element, rawVal]) => {
            const num = extractNumber(rawVal) || 0;
            if (num !== 0) {
              const clamped = Math.max(0, Math.min(1e4, num));
              map[element] = clamped;
              elementSet.add(element);
            }
          });
        }
        perTarget.push(map);
      });
      const elements = Array.from(elementSet);
      if (elements.length === 0) {
        try {
          showEmptyChartPlaceholder(ctx6.canvas.parentElement, "\u4ED8\u7740\u6642\u9593\u306E\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093");
        } catch (e) {
        }
        return;
      }
      let globalMax = 0;
      perTarget.forEach((pt) => {
        Object.values(pt).forEach((v) => {
          if (typeof v === "number" && Number.isFinite(v))
            globalMax = Math.max(globalMax, Math.abs(v));
        });
      });
      const toPercent = (v) => {
        if (!v || !Number.isFinite(v))
          return 0;
        if (globalMax <= 1.01)
          return v * 100;
        if (globalMax <= 100.5)
          return v;
        if (globalMax <= 1e4)
          return v / 1e4 * 100;
        return Math.max(0, Math.min(100, v));
      };
      const datasets = elements.map((el, ei) => {
        const data = perTarget.map((pt) => toPercent(pt[el] || 0));
        const hue = Math.abs(hashCode(el)) % 360;
        return {
          label: el,
          data,
          backgroundColor: `hsl(${hue}deg 70% 50%)`,
          stack: "Stack 0"
        };
      });
      const auraDesired = Math.max(200, targetLabels.length * 50);
      ensureContainerHeight(ctx6, auraDesired);
      try {
        setCanvasVisualSize(ctx6, auraDesired);
      } catch (e) {
      }
      charts.aura = new Chart(ctx6, {
        type: "bar",
        data: { labels: targetLabels, datasets },
        options: {
          // Horizontal bars: categories (targets) on Y axis, percent on X axis
          indexAxis: "y",
          plugins: {
            legend: { position: "top" },
            tooltip: {
              callbacks: {
                label: function(context) {
                  const raw = ctxRawValue(context);
                  return `${context.dataset.label || ""}: ${formatValue(raw, 2, "%")}`;
                }
              }
            }
          },
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            x: { stacked: true, beginAtZero: true, max: 100, ticks: { callback: function(v) {
              return v + "%";
            } } },
            y: { stacked: true, grid: { display: false } }
          }
        }
      });
      try {
        scheduleChartResize(charts.aura, ctx6);
      } catch (e) {
      }
      try {
        adjustAllChartInsets();
      } catch (e) {
      }
    })();
    console.log("[WebUI] Charts displayed, active charts:", Object.keys(charts));
  }
  function extractChartData(dataObj) {
    const labels = [];
    const values = [];
    if (typeof dataObj === "object" && dataObj !== null) {
      for (const [key, value] of Object.entries(dataObj)) {
        if (typeof value === "number") {
          if (!Number.isFinite(value))
            continue;
          labels.push(key);
          values.push(value);
        } else if (typeof value === "object" && value !== null && typeof value.mean === "number") {
          if (!Number.isFinite(value.mean))
            continue;
          labels.push(key);
          values.push(value.mean);
        }
      }
    }
    return { labels, values };
  }
  function showEmptyChartPlaceholder(containerEl, text) {
    try {
      if (!containerEl)
        return;
      const existing = containerEl.querySelector(".chart-empty-placeholder");
      if (existing)
        existing.remove();
      const div = document.createElement("div");
      div.className = "chart-empty-placeholder";
      div.style.padding = "24px";
      div.style.color = "var(--muted)";
      div.style.fontSize = "0.95rem";
      div.style.textAlign = "left";
      div.textContent = text || "\u30C7\u30FC\u30BF\u304C\u3042\u308A\u307E\u305B\u3093";
      containerEl.appendChild(div);
    } catch (e) {
    }
  }
  function createStackedBarChart(ctx, categories, [charNames, charValues, charSd], title) {
    const total = charValues.reduce((a, b) => a + b, 0);
    const percentages = charValues.map((v) => total > 0 ? v / total * 100 : 0);
    console.log("[WebUI] Calculated percentages:", percentages);
    const palette = CHAR_PALETTE;
    const datasets = charNames.map((name, idx) => {
      const hex = palette[idx % palette.length];
      const bg = hexToRgba(hex, 0.85);
      const border = hexToRgba(hex, 1);
      return {
        label: name,
        data: [percentages[idx]],
        stack: "stack1",
        backgroundColor: bg,
        borderColor: border,
        borderWidth: 1,
        hoverBackgroundColor: hexToRgba(hex, 0.95),
        // Make the bar thickness approximately 24px
        barThickness: 48,
        maxBarThickness: 48,
        categoryPercentage: 1,
        barPercentage: 1
      };
    });
    const numRows = Array.isArray(categories) && categories.length > 0 ? categories.length : 1;
    const barThickness = 48;
    const verticalPadding = 6;
    const legendSpace = 20;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);
    const chart = new Chart(ctx, {
      type: "bar",
      data: {
        labels: categories,
        datasets
      },
      options: {
        // Render horizontally: categories on the Y axis, values (percent) on the X axis
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        layout: { padding: { top: 0, bottom: 0, left: 0, right: 0 } },
        plugins: {
          legend: {
            display: true,
            position: "bottom",
            labels: { boxWidth: 12, padding: 4 }
          },
          title: {
            display: false
          },
          tooltip: {
            callbacks: {
              title: function() {
                return "";
              },
              label: function(context) {
                const charIdx = context.datasetIndex;
                const dps = charValues[charIdx] || 0;
                const sd = charSd && typeof charSd[charIdx] !== "undefined" && charSd[charIdx] !== null ? charSd[charIdx] : null;
                const pct = percentages[charIdx];
                const pctStr = pct.toFixed(2) + "%";
                const sdStr = sd === null ? "n/a" : sd.toFixed(2);
                const dpsStr = Number(dps).toLocaleString("ja-JP", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
                return `${context.dataset.label}: ${pctStr} (DPS: ${dpsStr} \xB1 ${sdStr})`;
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
              callback: function(value) {
                return value + "%";
              },
              padding: 4
            },
            grid: { drawBorder: false, display: false }
          },
          y: {
            stacked: true,
            display: false,
            grid: { display: false }
          }
        }
      }
    });
    try {
      chart.resize();
      chart.update();
    } catch (e) {
    }
    console.log("[WebUI] Chart created successfully, returning chart object");
    return chart;
  }
  function ensureContainerHeight(ctx, desiredHeightPx) {
    try {
      const canvas = ctx && ctx.canvas ? ctx.canvas : null;
      if (!canvas)
        return;
      const parent = canvas.parentElement;
      if (parent) {
        try {
          const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(parent).minHeight) : NaN;
          const existingInline = parent.style && parent.style.minHeight ? parseFloat(parent.style.minHeight) : NaN;
          const existing = !Number.isNaN(existingInline) && existingInline > 0 ? existingInline : Number.isNaN(computed) ? 0 : computed;
          const vp = typeof window !== "undefined" && window.innerHeight ? window.innerHeight : 800;
          const absoluteMax = Math.max(800, Math.floor(vp * 1.5));
          const target = Math.max(120, Math.min(desiredHeightPx, absoluteMax));
          if (target > existing + 4) {
            parent.style.setProperty("min-height", Math.ceil(target) + "px", "important");
          }
        } catch (e) {
        }
      }
      let el = parent;
      let depth = 0;
      while (el && depth < 4) {
        if (el.classList && el.classList.contains("col")) {
          try {
            const computed = window.getComputedStyle ? parseFloat(window.getComputedStyle(el).minHeight) : NaN;
            const existingInline = el.style && el.style.minHeight ? parseFloat(el.style.minHeight) : NaN;
            const existing = !Number.isNaN(existingInline) && existingInline > 0 ? existingInline : Number.isNaN(computed) ? 0 : computed;
            const vp = typeof window !== "undefined" && window.innerHeight ? window.innerHeight : 800;
            const absoluteMax = Math.max(800, Math.floor(vp * 1.5));
            const target = Math.max(120, Math.min(desiredHeightPx, absoluteMax));
            if (target > existing + 4)
              el.style.setProperty("min-height", Math.ceil(target) + "px", "important");
          } catch (e) {
          }
          break;
        }
        el = el.parentElement;
        depth++;
      }
    } catch (e) {
    }
  }
  function scheduleChartResize(chart, ctx, maxAttempts = 8) {
    try {
      let attempts = 0;
      const tryResize = () => {
        attempts++;
        const w = ctx && ctx.canvas ? ctx.canvas.offsetWidth : 0;
        if (w > 0 || attempts >= maxAttempts) {
          try {
            if (chart && typeof chart.resize === "function") {
              chart.resize();
              chart.update();
            }
          } catch (e) {
          }
        } else {
          setTimeout(tryResize, 120);
        }
      };
      setTimeout(tryResize, 120);
    } catch (e) {
    }
  }
  function createBarChart(ctx, labels, data, label, meta) {
    const palette = CHAR_PALETTE;
    const bgColors = labels.map((_, i) => hexToRgba(palette[i % palette.length], 0.75));
    const borderColors = labels.map((_, i) => hexToRgba(palette[i % palette.length], 1));
    const numRows = labels.length || 1;
    const barThickness = 48;
    const verticalPadding = 6;
    const legendSpace = 8;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);
    const datasets = [{
      label,
      data,
      backgroundColor: bgColors,
      borderColor: borderColors,
      borderWidth: 1,
      barThickness: 48,
      maxBarThickness: 48,
      categoryPercentage: 1,
      barPercentage: 0.9
    }];
    const chart = new Chart(ctx, {
      type: "bar",
      data: { labels, datasets },
      options: {
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        layout: { padding: { top: 0, bottom: 0, left: 0, right: 0 } },
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: function(context) {
                const idx = context.dataIndex;
                const val = data[idx] || 0;
                if (meta && meta[idx]) {
                  const m = meta[idx];
                  const mean = typeof m.mean === "number" ? m.mean : val;
                  const sd = typeof m.sd === "number" ? m.sd : null;
                  const min = typeof m.min === "number" ? m.min : null;
                  const max = typeof m.max === "number" ? m.max : null;
                  const sdStr = sd === null ? "n/a" : sd.toFixed(2);
                  return `${context.label}: ${mean.toFixed(2)} \xB1 ${sdStr}`;
                }
                return `${context.label}: ${typeof val === "number" ? val.toFixed(2) : val}`;
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
    try {
      chart.resize();
      chart.update();
    } catch (e) {
    }
    return chart;
  }
  function createLineChart(ctx, labels, data, label, options) {
    const opts = Object.assign({ heightPx: 140 }, options || {});
    setCanvasVisualSize(ctx, opts.heightPx);
    const chart = new Chart(ctx, {
      type: "line",
      data: {
        labels,
        datasets: [{
          label,
          data,
          backgroundColor: "rgba(102, 126, 234, 0.2)",
          borderColor: "rgba(102, 126, 234, 1)",
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
    try {
      chart.resize();
      chart.update();
    } catch (e) {
    }
    return chart;
  }
  function createStackedAbilitiesChart(ctx, charNames, abilities, matrix, title, metaMatrix, options) {
    const totalByAbility = abilities.map((ab, aIdx) => ({ idx: aIdx, total: matrix[aIdx].reduce((s, v) => s + (v || 0), 0) }));
    totalByAbility.sort((a, b) => b.total - a.total);
    const sortedAbilities = totalByAbility.map((t) => abilities[t.idx]);
    const sortedMatrix = totalByAbility.map((t) => matrix[t.idx]);
    const sortedMeta = metaMatrix ? totalByAbility.map((t) => metaMatrix[t.idx]) : null;
    const opts = Object.assign({ barThickness: 18, verticalPadding: 6 }, options || {});
    const datasets = charNames.map((char, cIdx) => {
      const hex = getCharColor(cIdx);
      const bg = hexToRgba(hex, 0.85);
      const border = hexToRgba(hex, 1);
      const data = sortedMatrix.map((row) => row[cIdx] || 0);
      return {
        label: char,
        data,
        stack: "stack1",
        backgroundColor: bg,
        borderColor: border,
        borderWidth: 1,
        barThickness: opts.barThickness,
        maxBarThickness: opts.barThickness,
        categoryPercentage: 1,
        barPercentage: 1
      };
    });
    const numRows = sortedAbilities.length || 1;
    const barThickness = opts.barThickness || 18;
    const verticalPadding = typeof opts.verticalPadding === "number" ? opts.verticalPadding : 6;
    const legendSpace = 8;
    const desiredHeightPx = Math.max(120, (barThickness + verticalPadding) * numRows + legendSpace);
    setCanvasVisualSize(ctx, desiredHeightPx);
    const chart = new Chart(ctx, {
      type: "bar",
      data: { labels: sortedAbilities, datasets },
      options: {
        indexAxis: "y",
        responsive: true,
        maintainAspectRatio: false,
        layout: { padding: { top: 0, bottom: 0, left: 0, right: 0 } },
        plugins: {
          legend: { display: true, position: "bottom", labels: { boxWidth: 12, padding: 4 } },
          tooltip: {
            callbacks: {
              title: function() {
                return "";
              },
              label: function(context) {
                const charIdx = context.datasetIndex;
                const abilityIdx = context.dataIndex;
                const ability = context.chart.data.labels[abilityIdx];
                const val = context.dataset.data[abilityIdx] || 0;
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
                return `${charNames[charIdx]}: ${ability}: ${typeof val === "number" ? val.toFixed(2) : val}`;
              }
            }
          }
        },
        scales: { x: { stacked: true, grid: { display: false }, ticks: { padding: 4 } }, y: { stacked: true, grid: { display: false }, ticks: { padding: 6 } } }
      }
    });
    try {
      chart.resize();
      chart.update();
    } catch (e) {
    }
    return chart;
  }
  function formatNumber(num) {
    if (num === void 0 || num === null)
      return "-";
    if (typeof num !== "number")
      return num;
    return Math.round(num).toLocaleString("ja-JP");
  }
  function formatStatWithStdev(mean, stdev) {
    if (mean === void 0 || mean === null)
      return "-";
    const meanFormatted = mean.toFixed(2);
    const stdevFormatted = stdev ? stdev.toFixed(2) : "0.00";
    return `${meanFormatted}<br><small style="font-size: 0.5em; font-weight: 400; color: #999;">\xB1${stdevFormatted}</small>`;
  }
  var _resizeTimer = null;
  window.addEventListener("resize", () => {
    if (_resizeTimer)
      clearTimeout(_resizeTimer);
    _resizeTimer = setTimeout(() => {
      adjustAllChartInsets();
    }, 150);
  });
  window.runSimulation = runSimulation;
  window.switchTab = switchTab;
  window.runOptimizerSimulation = runOptimizerSimulation;
})();
