<script lang="ts">
	import { connectionStatus, theme, toggleTheme, showMonitoringPanel, showControlPanel, toggleControlPanel } from '../../stores/elevators';

	function handleThemeToggle() {
		toggleTheme();
	}

	function toggleMonitoring() {
		showMonitoringPanel.update(show => !show);
	}

	function handleControlPanelToggle() {
		toggleControlPanel();
	}
</script>

<header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between h-16">
		<!-- Brand Identity -->
		<div class="flex items-center">
			<div class="flex-shrink-0">
				<h1 class="text-xl font-bold text-gray-900 dark:text-white">
					🏢 Elevator Control System
				</h1>
			</div>
		</div>

		<!-- Center - Connection Status -->
		<div class="flex items-center space-x-4">
			<div class="flex items-center">
				<div class="flex items-center space-x-2">
					<div class="h-3 w-3 rounded-full {$connectionStatus.connected ? 'bg-green-500' : 'bg-red-500'} {$connectionStatus.connected ? 'animate-pulse' : ''}"></div>
					<span class="text-sm font-medium text-gray-700 dark:text-gray-300">
						{$connectionStatus.connected ? 'Connected' : 'Disconnected'}
					</span>
					{#if $connectionStatus.retryCount > 0}
						<span class="text-xs text-gray-500 dark:text-gray-400">
							(Retry {$connectionStatus.retryCount})
						</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Right side - Controls -->
		<div class="flex items-center space-x-4">
			<!-- Control Panel Toggle -->
			<button
				type="button"
				class="control-panel-toggle text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors {$showControlPanel ? 'text-blue-600 dark:text-blue-400' : ''}"
				aria-label="Toggle control panel"
				on:click={handleControlPanelToggle}
			>
				{#if $showControlPanel}
					<!-- Show menu icon when panel is visible -->
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
					</svg>
				{:else}
					<!-- Show panel icon when panel is hidden -->
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
					</svg>
				{/if}
				<span class="tooltip">{$showControlPanel ? 'Hide Control Panel' : 'Show Control Panel'}</span>
			</button>

			<!-- Monitoring Panel Toggle -->
			<button
				type="button"
				class="monitoring-panel-toggle text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
				aria-label="Toggle monitoring panel"
				on:click={toggleMonitoring}
			>
				<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
				</svg>
				<span class="tooltip">{$showMonitoringPanel ? 'Hide Monitoring Panel' : 'Show Monitoring Panel'}</span>
			</button>

			<!-- Theme Toggle -->
			<button
				type="button"
				class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
				aria-label="Toggle theme"
				on:click={handleThemeToggle}
			>
				{#if $theme.mode === 'dark'}
					<!-- Sun Icon -->
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
					</svg>
				{:else}
					<!-- Moon Icon -->
					<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
					</svg>
				{/if}
			</button>

			<!-- Help Button -->
			<button
				type="button"
				class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
				aria-label="Help"
			>
				<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</button>
		</div>
	</div>
</header>

<style>
	/* Tooltip styles */
	.control-panel-toggle,
	.monitoring-panel-toggle {
		position: relative;
	}

	.tooltip {
		position: absolute;
		bottom: -40px;
		left: 50%;
		transform: translateX(-50%);
		background-color: #1f2937;
		color: white;
		padding: 8px 12px;
		border-radius: 6px;
		font-size: 12px;
		white-space: nowrap;
		opacity: 0;
		visibility: hidden;
		transition: opacity 0.2s ease, visibility 0.2s ease;
		z-index: 1000;
		pointer-events: none;
	}

	.tooltip::before {
		content: '';
		position: absolute;
		top: -4px;
		left: 50%;
		transform: translateX(-50%);
		border-left: 4px solid transparent;
		border-right: 4px solid transparent;
		border-bottom: 4px solid #1f2937;
	}

	.control-panel-toggle:hover .tooltip,
	.monitoring-panel-toggle:hover .tooltip {
		opacity: 1;
		visibility: visible;
	}

	/* Dark mode tooltip */
	@media (prefers-color-scheme: dark) {
		.tooltip {
			background-color: #374151;
		}

		.tooltip::before {
			border-bottom-color: #374151;
		}
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		h1 {
			font-size: 1.125rem;
		}
		
		.space-x-4 {
			gap: 0.5rem;
		}

		/* Hide tooltips on mobile for better UX */
		.tooltip {
			display: none;
		}
	}
</style> 