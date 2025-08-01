<script lang="ts">
	import '../styles/global.css';
	import { onMount } from 'svelte';
	import { theme } from '../stores/elevators';
	import { wsService } from '../services/websocket';
	import { elevatorAPI } from '../services/api';
	import { systemStatus } from '../stores/elevators';

	let isDark = false;

	onMount(() => {
		// Initialize theme from localStorage or system preference
		const savedTheme = localStorage.getItem('theme');
		if (savedTheme) {
			theme.set({ mode: savedTheme as 'light' | 'dark' });
			isDark = savedTheme === 'dark';
		} else {
			isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
			theme.set({ mode: isDark ? 'dark' : 'light' });
		}

		// Apply theme to document
		document.documentElement.classList.toggle('dark', isDark);

		// Listen for theme changes
		const unsubscribe = theme.subscribe(($theme) => {
			isDark = $theme.mode === 'dark';
			document.documentElement.classList.toggle('dark', isDark);
			localStorage.setItem('theme', $theme.mode);
		});

		// Initialize WebSocket connection
		wsService.connect();

		// Load initial data
		loadInitialData();

		// Set up periodic sync with backend (every 30 seconds)
		const syncInterval = setInterval(async () => {
			try {
				await elevatorAPI.syncElevatorStatus();
			} catch (error) {
				console.warn('Periodic sync failed:', error);
			}
		}, 30000);

		// Clean up interval on component destroy
		return () => {
			unsubscribe();
			wsService.disconnect();
			clearInterval(syncInterval);
		};
	});

	async function loadInitialData() {
		try {
			console.log('Loading initial data from backend...');

			// Sync current elevator status from backend
			await elevatorAPI.syncElevatorStatus();

			// Load system status
			const status = await elevatorAPI.getHealthStatus();
			if (status) {
				systemStatus.set(status);
				console.log('System status loaded:', status);
			} else {
				// Set default status if health check fails
				systemStatus.set({
					healthy: true,
					elevatorCount: 0,
					lastMaintenance: undefined,
					alerts: [],
				});
			}
		} catch (error) {
			console.error('Failed to load initial data:', error);
			// Set default status on error
			systemStatus.set({
				healthy: true,
				elevatorCount: 0,
				lastMaintenance: undefined,
				alerts: [],
			});
		}
	}
</script>

<svelte:head>
	<title>Elevator Control System</title>
	<meta
		name="description"
		content="Modern Elevator Control System - Real-time visualization and control interface"
	/>
</svelte:head>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors duration-300">
	<slot />
</div>

<style>
	:global(html) {
		height: 100%;
	}

	:global(body) {
		height: 100%;
		margin: 0;
		padding: 0;
	}
</style>
