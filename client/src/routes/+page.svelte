<script lang="ts">
	import Header from '../components/layout/Header.svelte';
	import ElevatorControlPanel from '../components/controls/ElevatorControlPanel.svelte';
	import ElevatorBuildingGrid from '../components/elevator/ElevatorBuildingGrid.svelte';
	import MonitoringDashboard from '../components/monitoring/MonitoringDashboard.svelte';
	import Footer from '../components/layout/Footer.svelte';
	import Toast from '../components/common/Toast.svelte';
	import { showMonitoringPanel, showControlPanel } from '../stores/elevators';
</script>

<div class="flex flex-col min-h-screen">
	<!-- Header -->
	<Header />

	<!-- Main Content -->
	<div class="flex-1 flex flex-col lg:flex-row">
		<!-- Left Sidebar - Control Panel -->
		{#if $showControlPanel}
			<div class="w-full lg:w-80 xl:w-96 border-r border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 transition-all duration-300 ease-in-out">
				<ElevatorControlPanel />
			</div>
		{/if}

		<!-- Main Content Area - Elevator Visualization -->
		<div class="flex-1 min-h-0">
			<ElevatorBuildingGrid />
		</div>

		<!-- Right Sidebar - Monitoring Dashboard (collapsible) -->
		{#if $showMonitoringPanel}
			<div class="w-full lg:w-80 xl:w-96 border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
				<MonitoringDashboard />
			</div>
		{/if}
	</div>

	<!-- Footer -->
	<Footer />

	<!-- Toast Notifications -->
	<Toast />
</div>

<style>
	/* Responsive adjustments */
	@media (max-width: 768px) {
		.flex-col {
			flex-direction: column;
		}
		
		.w-full.lg\:w-80 {
			width: 100%;
		}
	}
</style> 