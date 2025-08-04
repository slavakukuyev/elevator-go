<script lang="ts">
	import {
		elevators,
		systemStatus,
		connectionStatus,
		elevatorUtilization,
		pendingRequests,
		addNotification,
	} from '../../stores/elevators';
	import { elevatorAPI } from '../../services/api';

	$: healthyElevators = $elevators.filter((e) => e.status !== 'error').length;
	$: errorElevators = $elevators.filter((e) => e.status === 'error').length;
	$: movingElevators = $elevators.filter((e) => e.status === 'moving').length;
	$: idleElevators = $elevators.filter((e) => e.status === 'idle').length;

	async function refreshData() {
		try {
			await elevatorAPI.syncElevatorStatus();
			addNotification('Data refreshed successfully');
		} catch (error) {
			console.error('Failed to refresh data:', error);
			addNotification('Failed to refresh data');
		}
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">System Monitoring</h2>
		<button
			on:click={refreshData}
			class="bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 text-white px-3 py-1 rounded text-sm font-medium transition-colors flex items-center gap-2"
		>
			<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
				/>
			</svg>
			Refresh
		</button>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto p-4 space-y-6">
		<!-- Connection Status -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
		>
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Connection Status</h3>
			<div class="flex items-center space-x-3">
				<div class="flex items-center">
					<div
						class="h-3 w-3 rounded-full {$connectionStatus.connected
							? 'bg-green-500'
							: 'bg-red-500'} mr-2"
					/>
					<span class="text-sm text-gray-600 dark:text-gray-400">
						{$connectionStatus.connected ? 'Connected' : 'Disconnected'}
					</span>
				</div>
				{#if $connectionStatus.retryCount > 0}
					<span class="text-xs text-gray-500 dark:text-gray-400">
						Retry: {$connectionStatus.retryCount}
					</span>
				{/if}
			</div>
		</div>

		<!-- System Health -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
		>
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">System Health</h3>
			<div class="space-y-3">
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Overall Status</span>
					<span
						class="text-sm font-medium {$systemStatus.healthy
							? 'text-green-600 dark:text-green-400'
							: 'text-red-600 dark:text-red-400'}"
					>
						{$systemStatus.healthy ? 'Healthy' : 'Issues Detected'}
					</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Total Elevators</span>
					<span class="text-sm font-medium text-gray-900 dark:text-white"
						>{$systemStatus.elevatorCount}</span
					>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Operational</span>
					<span class="text-sm font-medium text-green-600 dark:text-green-400"
						>{healthyElevators}</span
					>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Errors</span>
					<span class="text-sm font-medium text-red-600 dark:text-red-400">{errorElevators}</span>
				</div>
			</div>
		</div>

		<!-- Elevator Status Breakdown -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
		>
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Elevator Status</h3>
			<div class="space-y-3">
				<div class="flex items-center justify-between">
					<div class="flex items-center">
						<div class="h-2 w-2 rounded-full bg-green-500 mr-2" />
						<span class="text-sm text-gray-600 dark:text-gray-400">Idle</span>
					</div>
					<span class="text-sm font-medium text-gray-900 dark:text-white">{idleElevators}</span>
				</div>
				<div class="flex items-center justify-between">
					<div class="flex items-center">
						<div class="h-2 w-2 rounded-full bg-blue-500 mr-2" />
						<span class="text-sm text-gray-600 dark:text-gray-400">Moving</span>
					</div>
					<span class="text-sm font-medium text-gray-900 dark:text-white">{movingElevators}</span>
				</div>
				<div class="flex items-center justify-between">
					<div class="flex items-center">
						<div class="h-2 w-2 rounded-full bg-red-500 mr-2" />
						<span class="text-sm text-gray-600 dark:text-gray-400">Error</span>
					</div>
					<span class="text-sm font-medium text-gray-900 dark:text-white">{errorElevators}</span>
				</div>
			</div>
		</div>

		<!-- Utilization -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
		>
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Utilization</h3>
			<div class="space-y-3">
				<div>
					<div class="flex justify-between text-sm mb-1">
						<span class="text-gray-600 dark:text-gray-400">Idle</span>
						<span class="text-gray-900 dark:text-white"
							>{$elevatorUtilization.idle.toFixed(1)}%</span
						>
					</div>
					<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
						<div
							class="bg-green-500 h-2 rounded-full"
							style="width: {$elevatorUtilization.idle}%"
						/>
					</div>
				</div>
				<div>
					<div class="flex justify-between text-sm mb-1">
						<span class="text-gray-600 dark:text-gray-400">Active</span>
						<span class="text-gray-900 dark:text-white"
							>{$elevatorUtilization.moving.toFixed(1)}%</span
						>
					</div>
					<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
						<div
							class="bg-blue-500 h-2 rounded-full"
							style="width: {$elevatorUtilization.moving}%"
						/>
					</div>
				</div>
				<div>
					<div class="flex justify-between text-sm mb-1">
						<span class="text-gray-600 dark:text-gray-400">Error</span>
						<span class="text-gray-900 dark:text-white"
							>{$elevatorUtilization.error.toFixed(1)}%</span
						>
					</div>
					<div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
						<div class="bg-red-500 h-2 rounded-full" style="width: {$elevatorUtilization.error}%" />
					</div>
				</div>
			</div>
		</div>

		<!-- Pending Requests -->
		<div
			class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
		>
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">
				Pending Requests ({$pendingRequests.length})
			</h3>
			{#if $pendingRequests.length > 0}
				<div class="space-y-2 max-h-48 overflow-y-auto">
					{#each $pendingRequests.slice(0, 10) as request}
						<div
							class="flex items-center justify-between py-2 px-3 bg-gray-50 dark:bg-gray-700 rounded"
						>
							<div class="text-sm">
								<span class="text-gray-900 dark:text-white">{request.from} → {request.to}</span>
							</div>
							<div class="text-xs text-gray-500 dark:text-gray-400">
								{new Date(request.timestamp).toLocaleTimeString()}
							</div>
						</div>
					{/each}
					{#if $pendingRequests.length > 10}
						<div class="text-xs text-center text-gray-500 dark:text-gray-400 py-2">
							+{$pendingRequests.length - 10} more
						</div>
					{/if}
				</div>
			{:else}
				<div class="text-center py-4">
					<div class="text-gray-400 text-sm">No pending requests</div>
				</div>
			{/if}
		</div>

		<!-- Individual Elevator Details -->
		{#if $elevators.length > 0}
			<div
				class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700"
			>
				<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">Elevator Details</h3>
				<div class="space-y-3">
					{#each $elevators as elevator}
						<div class="border border-gray-200 dark:border-gray-700 rounded-lg p-3">
							<div class="flex items-center justify-between mb-2">
								<span class="text-sm font-medium text-gray-900 dark:text-white"
									>{elevator.name}</span
								>
								<div class="flex items-center space-x-2">
									<div
										class="h-2 w-2 rounded-full {elevator.status === 'idle'
											? 'bg-green-500'
											: elevator.status === 'moving'
											? 'bg-blue-500'
											: 'bg-red-500'}"
									/>
									<span class="text-xs text-gray-600 dark:text-gray-400 capitalize"
										>{elevator.status}</span
									>
								</div>
							</div>
							<div class="text-xs text-gray-600 dark:text-gray-400 space-y-1">
								<div class="flex justify-between">
									<span>Current Floor:</span>
									<span>{elevator.currentFloor}</span>
								</div>
								<div class="flex justify-between">
									<span>Range:</span>
									<span>{elevator.minFloor} - {elevator.maxFloor}</span>
								</div>
								{#if elevator.direction}
									<div class="flex justify-between">
										<span>Direction:</span>
										<span class="capitalize">{elevator.direction}</span>
									</div>
								{/if}
								<div class="flex justify-between">
									<span>Doors:</span>
									<span>{elevator.doorsOpen ? 'Open' : 'Closed'}</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	/* Custom scrollbar for pending requests */
	.overflow-y-auto::-webkit-scrollbar {
		width: 4px;
	}

	.overflow-y-auto::-webkit-scrollbar-track {
		background: transparent;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb {
		background: rgba(156, 163, 175, 0.5);
		border-radius: 2px;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb:hover {
		background: rgba(156, 163, 175, 0.7);
	}
</style>
