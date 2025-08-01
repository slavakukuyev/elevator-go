<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { floorSelectionService } from '../../utils/floorSelection';
	import { elevatorAPI } from '../../services/api';
	import { elevators } from '../../stores/elevators';
	import Button from '../common/Button.svelte';

	export let floor: number;
	export let compact: boolean = false;

	const dispatch = createEventDispatcher<{
		floorSelected: { from: number; to: number; assignedElevator: string | null };
	}>();

	let selectedDestination: number | null = null;
	let isRequesting = false;
	let assignedElevator: string | null = null;
	let showAssignment = false;
	let assignmentTimer: NodeJS.Timeout | null = null;

	// Calculate available destinations based on current floor and available elevators
	$: availableDestinations = floorSelectionService.getAvailableDestinations(floor, $elevators);

	function formatFloor(floorNum: number): string {
		return floorSelectionService.formatFloorDisplay(floorNum);
	}

	function getFloorDisplayClass(floorNum: number): string {
		if (floorNum === 0)
			return 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-600';
		if (floorNum < 0) return 'bg-blue-100 dark:bg-blue-900/20 border-blue-300 dark:border-blue-600';
		return 'bg-gray-100 dark:bg-gray-800 border-gray-300 dark:border-gray-600';
	}

	async function handleDestinationSelect(destination: number) {
		if (isRequesting || destination === floor) return;

		selectedDestination = destination;
		isRequesting = true;
		assignedElevator = null;
		showAssignment = false;

		try {
			await elevatorAPI.requestFloor(floor, destination);

			// Find the optimal elevator for this request
			const optimalElevator = floorSelectionService.findOptimalElevator(
				floor,
				destination,
				$elevators
			);

			assignedElevator = optimalElevator?.name || null;
			showAssignment = true;

			// Dispatch event to parent
			dispatch('floorSelected', {
				from: floor,
				to: destination,
				assignedElevator,
			});

			// Auto-hide assignment after 5 seconds
			if (assignmentTimer) clearTimeout(assignmentTimer);
			assignmentTimer = setTimeout(() => {
				showAssignment = false;
				selectedDestination = null;
				assignedElevator = null;
			}, 5000);
		} catch (error) {
			console.error('Failed to request floor:', error);
		} finally {
			isRequesting = false;
		}
	}

	function clearAssignment() {
		showAssignment = false;
		selectedDestination = null;
		assignedElevator = null;
		if (assignmentTimer) clearTimeout(assignmentTimer);
	}
</script>

<div
	class="relative bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm {compact
		? 'p-3'
		: 'p-4'}"
>
	{#if showAssignment}
		<!-- Assignment Response -->
		<div
			class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4 mb-4"
		>
			<div class="flex items-center justify-between">
				<div class="flex items-center">
					<svg
						class="h-5 w-5 text-green-600 dark:text-green-400 mr-2"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 13l4 4L19 7"
						/>
					</svg>
					<div>
						<div class="text-sm font-medium text-green-900 dark:text-green-100">
							{#if assignedElevator}
								Wait for Elevator {assignedElevator}
							{:else}
								Request Processed
							{/if}
						</div>
						<div class="text-xs text-green-700 dark:text-green-300">
							{formatFloor(floor)} → {formatFloor(selectedDestination ?? 0)}
						</div>
					</div>
				</div>
				<Button
					variant="outline"
					size="small"
					on:click={clearAssignment}
					ariaLabel="Clear assignment"
				>
					✕
				</Button>
			</div>
		</div>
	{/if}

	<!-- Floor Selection Header -->
	{#if !compact}
		<div class="mb-4">
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-1">
				Floor {formatFloor(floor)}
			</h3>
			<p class="text-sm text-gray-600 dark:text-gray-400">Select your destination floor</p>
		</div>
	{/if}

	<!-- Available Destinations -->
	{#if availableDestinations.length > 0}
		<div class="grid gap-2 {compact ? 'grid-cols-4' : 'grid-cols-6'}">
			{#each availableDestinations as destination}
				<Button
					variant="outline"
					size={compact ? 'small' : 'medium'}
					class="aspect-square hover:scale-105 transition-transform duration-200 {compact
						? 'min-h-[40px] text-sm'
						: 'min-h-[60px]'} {getFloorDisplayClass(destination)} {selectedDestination ===
					destination
						? 'ring-2 ring-blue-500'
						: ''}"
					on:click={() => handleDestinationSelect(destination)}
					disabled={isRequesting || destination === floor}
					ariaLabel="Go to floor {formatFloor(destination)}"
				>
					<div class="text-center">
						<div class="font-bold text-gray-900 dark:text-white">
							{formatFloor(destination)}
						</div>
						{#if !compact}
							{#if destination === 0}
								<div class="text-xs text-gray-600 dark:text-gray-400">Ground</div>
							{:else if destination < 0}
								<div class="text-xs text-gray-600 dark:text-gray-400">Basement</div>
							{/if}
						{/if}
					</div>
				</Button>
			{/each}
		</div>
	{:else}
		<div class="text-center py-4">
			<p class="text-sm text-gray-600 dark:text-gray-400">No destinations available</p>
		</div>
	{/if}

	<!-- Loading State -->
	{#if isRequesting}
		<div
			class="absolute inset-0 bg-white/80 dark:bg-gray-900/80 flex items-center justify-center rounded-lg z-10"
		>
			<div class="text-center">
				<div class="spinner mb-2" aria-hidden="true" />
				<p class="text-sm text-gray-600 dark:text-gray-400">Processing request...</p>
			</div>
		</div>
	{/if}
</div>
