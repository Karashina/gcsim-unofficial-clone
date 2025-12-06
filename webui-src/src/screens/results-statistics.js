/**
 * Results Screen - Statistics Summary Module
 * Displays DPS, EPS, RPS, HPS, SHP, Duration with standard deviations
 */

/**
 * Display statistics summary
 * @param {Object} result - Simulation result object
 */
export function displayStatistics(result) {
    console.log('[WebUI] Displaying statistics...');
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
    
    console.log('[WebUI] Stats:', { dps, eps, rps, hps, shp, duration });
    
    // Display each stat
    updateStat('stat-dps', dps, dpsStd);
    updateStat('stat-eps', eps, epsStd);
    updateStat('stat-rps', rps, rpsStd);
    updateStat('stat-hps', hps, hpsStd);
    updateStat('stat-shp', shp, shpStd);
    updateStat('stat-dur', duration, durationStd);
}

/**
 * Update a single stat display
 * @param {string} elementId - Element ID
 * @param {number} value - Main value
 * @param {number} stdev - Standard deviation
 */
function updateStat(elementId, value, stdev) {
    const element = document.getElementById(elementId);
    if (!element) {
        console.warn(`[WebUI] Element ${elementId} not found`);
        return;
    }
    
    element.innerHTML = formatStatWithStdev(value, stdev);
}

/**
 * Format statistic with standard deviation
 * @param {number} value - Main value
 * @param {number} stdev - Standard deviation
 * @returns {string} Formatted HTML
 */
function formatStatWithStdev(value, stdev) {
    const formattedValue = Number(value).toLocaleString('ja-JP', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    });
    
    if (stdev && stdev > 0) {
        const formattedStdev = Number(stdev).toLocaleString('ja-JP', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2
        });
        return `${formattedValue}<br><span class="stat-stdev">±${formattedStdev}</span>`;
    }
    
    return formattedValue;
}
