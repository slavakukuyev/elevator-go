<script lang="ts">
	import Button from '../common/Button.svelte';
	import CreateElevatorModal from './CreateElevatorModal.svelte';
	import { systemStatus, elevators } from '../../stores/elevators';
	import { floorSelectionService } from '../../utils/floorSelection';
	import { elevatorAPI } from '../../services/api';
	import type { Elevator } from '../../types';

	let createModalOpen = false;
	let selectedFloor = 0;

	$: totalFloors =
		$elevators.length > 0
			? Math.max(...$elevators.map((e: Elevator) => e.maxFloor ?? 0)) -
			  Math.min(...$elevators.map((e: Elevator) => e.minFloor ?? 0)) +
			  1
			: 0;

	$: availableFloors =
		$elevators.length > 0
			? Array.from(
					{
						length:
							Math.max(...$elevators.map((e: Elevator) => e.maxFloor ?? 0)) -
							Math.min(...$elevators.map((e: Elevator) => e.minFloor ?? 0)) +
							1,
					},
					(_, i) => Math.min(...$elevators.map((e: Elevator) => e.minFloor ?? 0)) + i
			  )
			: [];

	// Generate quick access floors based on available elevators from current floor
	$: quickAccessFloors =
		$elevators.length > 0
			? floorSelectionService.getAvailableDestinations(selectedFloor, $elevators)
			: [];

	function openCreateModal() {
		createModalOpen = true;
	}

	async function handleFloorRequest(toFloor: number) {
		try {
			// Validate that we have elevators available
			if ($elevators.length === 0) {
				console.error('No elevators available');
				return;
			}

			// Validate the floor request using our floor selection service
			const validation = floorSelectionService.validateFloorRequest(
				selectedFloor,
				toFloor,
				$elevators
			);
			if (!validation.valid) {
				console.error('Invalid floor request:', validation.message);
				return;
			}

			await elevatorAPI.requestFloor(selectedFloor, toFloor);
		} catch (error) {
			console.error('Failed to request floor:', error);
		}
	}

	async function handleDirectFloorRequest() {
		try {
			await elevatorAPI.requestFloor(0, 5);
		} catch (error) {
			console.error('Failed to request floor 0 to 5:', error);
		}
	}

	async function refreshStatus() {
		try {
			await elevatorAPI.syncElevatorStatus();
		} catch (error) {
			console.error('Failed to refresh status:', error);
		}
	}

	function formatFloor(floor: number | undefined): string {
		return floorSelectionService.formatFloorDisplay(floor);
	}
</script>

<div class="h-full flex flex-col">
	<!-- Header -->
	<div class="p-4 border-b border-gray-200 dark:border-gray-700">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Control Panel</h2>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto p-4 space-y-6">
		<!-- System Status -->
		<div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
			<h3 class="text-sm font-medium text-gray-900 dark:text-white mb-3">System Status</h3>
			<div class="space-y-2">
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Health</span>
					<span
						class="text-sm font-medium {$systemStatus.healthy
							? 'text-green-600 dark:text-green-400'
							: 'text-red-600 dark:text-red-400'}"
					>
						{$systemStatus.healthy ? 'Healthy' : 'Issues'}
					</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Elevators</span>
					<span class="text-sm font-medium text-gray-900 dark:text-white"
						>{$systemStatus.elevatorCount}</span
					>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-sm text-gray-600 dark:text-gray-400">Total Floors</span>
					<span class="text-sm font-medium text-gray-900 dark:text-white">{totalFloors}</span>
				</div>
			</div>
			<Button
				variant="secondary"
				size="small"
				fullWidth
				on:click={refreshStatus}
				ariaLabel="Refresh elevator status"
			>
				<svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8 0 004.582 9H4m0 0v5h4.582M10 18h4"
					/>
				</svg>
				Refresh Status
			</Button>
		</div>

		<!-- Create Elevator -->
		<div class="space-y-3">
			<h3 class="text-sm font-medium text-gray-900 dark:text-white">Elevator Management</h3>
			<Button
				variant="primary"
				size="medium"
				fullWidth
				on:click={openCreateModal}
				ariaLabel="Create new elevator"
			>
				<svg class="h-5 w-5 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 6v6m0 0v6m0-6h6m-6 0H6"
					/>
				</svg>
				Add Elevator
			</Button>
		</div>

		<!-- Floor Selection -->
		{#if $elevators.length > 0}
			<div class="space-y-3">
				<h3 class="text-sm font-medium text-gray-900 dark:text-white">Floor Controls</h3>

				<!-- Current Floor Selector -->
				<div>
					<label
						for="current-floor"
						class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
					>
						Current Floor
					</label>
					<select
						id="current-floor"
						bind:value={selectedFloor}
						class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white"
					>
						{#each availableFloors as floor}
							<option value={floor}>{formatFloor(floor)}</option>
						{/each}
					</select>
				</div>

				<!-- Quick Floor Access -->
				<div class="space-y-2">
					<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Quick Access</h4>
					<div class="grid grid-cols-3 gap-2">
						{#each quickAccessFloors as floor}
							<Button
								variant="secondary"
								size="small"
								on:click={() => handleFloorRequest(floor)}
								disabled={floor === selectedFloor}
								ariaLabel="Go to floor {formatFloor(floor)}"
							>
								{formatFloor(floor)}
							</Button>
						{/each}
					</div>
				</div>

				<!-- Direct Floor Request -->
				<div class="space-y-2">
					<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">Direct Request</h4>
					<Button
						variant="primary"
						size="medium"
						fullWidth
						on:click={handleDirectFloorRequest}
						ariaLabel="Move elevator from floor 0 to floor 5"
					>
						<svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M7 11l5-5m0 0l5 5m-5-5v12"
							/>
						</svg>
						Move 0→5
					</Button>
				</div>
			</div>
		{:else}
			<!-- No Elevators State -->
			<div class="text-center py-8">
				<svg
					class="h-12 w-12 mx-auto text-gray-400"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
					/>
				</svg>
				<h3 class="mt-2 text-sm font-medium text-gray-900 dark:text-white">No elevators</h3>
				<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
					Get started by creating your first elevator.
				</p>
			</div>
		{/if}
	</div>
</div>

<!-- Create Elevator Modal -->
<CreateElevatorModal bind:open={createModalOpen} />

<style>
	/* Custom scrollbar */
	.overflow-y-auto::-webkit-scrollbar {
		width: 6px;
	}

	.overflow-y-auto::-webkit-scrollbar-track {
		background: transparent;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb {
		background: rgba(156, 163, 175, 0.5);
		border-radius: 3px;
	}

	.overflow-y-auto::-webkit-scrollbar-thumb:hover {
		background: rgba(156, 163, 175, 0.7);
	}
</style>
