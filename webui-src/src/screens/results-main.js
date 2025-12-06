/**
 * Results Screen - Main Orchestrator
 * Coordinates display of all results sections
 */

import { displayCharacters } from './results-characters.js';
import { displayStatistics } from './results-statistics.js';
import { displayTargets } from './results-targets.js';
import { displayCharts } from './results-charts.js';

/**
 * Display all simulation results
 * @param {Object} result - Simulation result object
 */
export function displayResults(result) {
    console.log('[WebUI] Displaying results...');
    
    // Ensure resultsContainer is available (fix for TDZ issue)
    const resultsContainer = document.getElementById('results-container');
    if (!resultsContainer) {
        console.error('[WebUI] results-container not found');
        return;
    }
    
    // Make sure the results screen is visible
    const resultsScreen = document.getElementById('screen-results');
    if (resultsScreen && !resultsScreen.classList.contains('active')) {
        const resultsTab = document.querySelector('.navbar-tab[data-screen="results"]');
        if (resultsTab) {
            resultsTab.click();
        }
    }
    
    // Display each section
    displayTargets(result);
    displayCharacters(result);
    displayStatistics(result);
    displayCharts(result, resultsContainer);
    
    // Make results visible after layout is complete
    setTimeout(() => {
        try {
            resultsContainer.style.display = 'block';
            resultsContainer.classList.add('visible');
            resultsContainer.scrollIntoView({ behavior: 'smooth' });
        } catch (e) {
            console.warn('[WebUI] Could not scroll to results', e);
        }
    }, 100);
    
    console.log('[WebUI] Results displayed successfully');
}
