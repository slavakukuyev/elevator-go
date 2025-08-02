<script lang="ts">
	import {
		isLoading,
		elevators,
		showCreateModal,
		initializeSampleData,
	} from '../../stores/elevators';
	import type { Elevator } from '../../types';
	import ElevatorBuilding from './ElevatorBuilding.svelte';
	import CreateElevatorModal from '../controls/CreateElevatorModal.svelte';

	// Group elevators by type (main vs parking)
	$: mainElevators = $elevators.filter(
		(e: Elevator) => e.name && !e.name.toLowerCase().includes('parking')
	);
	$: parkingElevators = $elevators.filter(
		(e: Elevator) => e.name && e.name.toLowerCase().includes('parking')
	);

	function handleCreateElevator() {
		showCreateModal.set(true);
	}
</script>

<div class="h-full bg-gray-50 dark:bg-gray-900 overflow-auto">
	{#if $isLoading}
		<!-- Loading State -->
		<div class="h-full flex items-center justify-center">
			<div class="text-center">
				<div class="spinner large mb-4" aria-hidden="true" />
				<p class="text-lg text-gray-600 dark:text-gray-400">Loading building...</p>
			</div>
		</div>
	{:else if $elevators.length === 0}
		<!-- Empty State -->
		<div class="h-full flex items-center justify-center">
			<div class="text-center max-w-md mx-auto px-4">
				<div
					class="bg-gray-100 dark:bg-gray-800 rounded-full w-24 h-24 flex items-center justify-center mx-auto mb-6"
				>
					<svg
						class="h-12 w-12 text-gray-400"
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
				</div>
				<h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">Building Ready</h2>
				<p class="text-gray-600 dark:text-gray-400 mb-6">
					Create your first elevator to start the elevator control system.
				</p>
				<div class="flex gap-4">
					<button
						on:click={handleCreateElevator}
						class="bg-blue-600 hover:bg-blue-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
					>
						+ Create Elevator
					</button>
					<button
						on:click={() => initializeSampleData()}
						class="bg-gray-600 hover:bg-gray-700 text-white px-6 py-3 rounded-lg font-medium transition-colors"
					>
						Load Sample Data
					</button>
				</div>
			</div>
		</div>
	{:else}
		<!-- Elevator Building Grid -->
		<div class="elevator-building-grid p-6">
			<!-- Header -->
			<div class="mb-6">
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Elevator Control System</h1>
			</div>

			<!-- Main Elevators Section -->
			{#if mainElevators.length > 0}
				<div class="mb-8">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Main Elevators</h2>
					<div class="grid">
						{#each mainElevators as elevator (elevator.name)}
							<ElevatorBuilding {elevator} />
						{/each}
					</div>
				</div>
			{/if}

			<!-- Parking Elevators Section -->
			{#if parkingElevators.length > 0}
				<div>
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">
						Parking Elevators
					</h2>
					<div class="grid">
						{#each parkingElevators as elevator (elevator.name)}
							<ElevatorBuilding {elevator} />
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create Elevator Modal -->
{#if $showCreateModal}
	<CreateElevatorModal />
{/if}

<style>
	/* Ensure proper scrolling behavior */
	.h-full {
		overflow: hidden;
	}

	.elevator-building-grid {
		min-height: 100%;
	}

	/* Responsive grid adjustments */
	@media (max-width: 768px) {
		.elevator-building-grid {
			padding: 1rem;
		}
	}
</style>
