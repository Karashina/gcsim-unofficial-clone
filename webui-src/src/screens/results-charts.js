/**
 * Results Screen - Charts Module (Simplified Stub)
 * This is a simplified version - full chart rendering depends on Chart.js availability
 */

/**
 * Display charts for simulation results
 * @param {Object} result - Simulation result object
 * @param {HTMLElement} resultsContainer - Results container element
 */
export function displayCharts(result, resultsContainer) {
    console.log('[WebUI] Displaying charts...');
    
    const stats = result.statistics || {};
    
    // Debug panel removed: do not inject raw JSON into the UI
    
    // Check if Chart.js is available
    if (typeof Chart === 'undefined') {
        console.warn('[WebUI] Chart.js not available, skipping chart rendering');
        showChartPlaceholders();
        return;
    }
    
    // Render individual charts
    try {
        renderCharacterDPSChart(result, stats);
    } catch (e) {
        console.error('[WebUI] Error rendering character DPS chart:', e);
    }
    
    try {
        renderSourceDPSChart(result, stats);
    } catch (e) {
        console.error('[WebUI] Error rendering source DPS chart:', e);
    }
}

// Note: Raw statistics debug panel removed to prevent exposing large JSON in the UI.

/**
 * Show placeholder messages when Chart.js is not available
 */
function showChartPlaceholders() {
    const chartIds = ['char-dps-chart', 'source-dps-chart', 'damage-dist-chart'];
    
    chartIds.forEach(id => {
        const canvas = document.getElementById(id);
        if (canvas && canvas.parentElement) {
            const container = canvas.parentElement;
            const placeholder = document.createElement('div');
            placeholder.className = 'chart-empty';
            placeholder.textContent = 'グラフを表示するにはChart.jsが必要です';
            container.appendChild(placeholder);
            canvas.style.display = 'none';
        }
    });
}

/**
 * Render character DPS pie chart
 * @param {Object} result - Simulation result
 * @param {Object} stats - Statistics object
 */
function renderCharacterDPSChart(result, stats) {
    const canvas = document.getElementById('char-dps-chart');
    if (!canvas) {
        console.error('[WebUI] Canvas element char-dps-chart not found');
        return;
    }
    
    // Ensure canvas is visible (fix for visibility bug)
    canvas.style.visibility = 'visible';
    canvas.style.display = 'block';
    
    // Chart rendering logic would go here
    // For now, just log that we found the canvas
    console.log('[WebUI] Character DPS chart canvas ready');
}

/**
 * Render source DPS horizontal bar chart
 * @param {Object} result - Simulation result  
 * @param {Object} stats - Statistics object
 */
function renderSourceDPSChart(result, stats) {
    const canvas = document.getElementById('source-dps-chart');
    if (!canvas) {
        console.error('[WebUI] Canvas element source-dps-chart not found');
        return;
    }
    
    // Ensure canvas is visible (fix for visibility bug)
    canvas.style.visibility = 'visible';
    canvas.style.display = 'block';
    
    // Chart rendering logic would go here
    console.log('[WebUI] Source DPS chart canvas ready');
}
