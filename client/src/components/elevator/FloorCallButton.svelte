<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { floorSelectionService } from '../../utils/floorSelection';
	import { elevatorAPI } from '../../services/api';
	import { elevators } from '../../stores/elevators';
	import Modal from '../common/Modal.svelte';
	import Button from '../common/Button.svelte';

	export let floor: number;

	const dispatch = createEventDispatcher<{
		floorSelected: { from: number; to: number; assignedElevator: string | null };
	}>();

	let showModal = false;
	let showTooltip = false;
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

	function handleButtonClick() {
		if (availableDestinations.length > 0) {
			showModal = true;
		}
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
			showModal = false;
		}
	}

	function clearAssignment() {
		showAssignment = false;
		selectedDestination = null;
		assignedElevator = null;
		if (assignmentTimer) clearTimeout(assignmentTimer);
	}

	function handleModalClose() {
		showModal = false;
	}
</script>

<div class="relative">
	<!-- Metallic Call Button -->
	<button
		class="floor-call-button w-12 h-12 rounded-full bg-gradient-to-br from-gray-300 to-gray-500 dark:from-gray-600 dark:to-gray-800
		       border-2 border-gray-400 dark:border-gray-600 shadow-lg hover:shadow-xl
		       transition-all duration-200 hover:scale-110 active:scale-95
		       focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
		       disabled:opacity-50 disabled:cursor-not-allowed"
		disabled={availableDestinations.length === 0 || isRequesting}
		on:click={handleButtonClick}
		on:mouseenter={() => (showTooltip = true)}
		on:mouseleave={() => (showTooltip = false)}
		on:focus={() => (showTooltip = true)}
		on:blur={() => (showTooltip = false)}
		aria-label="Request elevator from floor {formatFloor(floor)}"
	>
		<!-- Button Icon -->
		<svg
			class="w-6 h-6 text-gray-700 dark:text-gray-300 mx-auto"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M7 11l5-5m0 0l5 5m-5-5v12"
			/>
		</svg>
	</button>

	<!-- Tooltip -->
	{#if showTooltip && availableDestinations.length > 0}
		<div
			class="absolute bottom-full left-1/2 transform -translate-x-1/2 mb-2 px-3 py-1
			       bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-sm rounded-md
			       shadow-lg z-10 whitespace-nowrap"
			role="tooltip"
		>
			Request elevator
			<!-- Tooltip arrow -->
			<div
				class="absolute top-full left-1/2 transform -translate-x-1/2 -mt-1
				       border-4 border-transparent border-t-gray-900 dark:border-t-gray-100"
			/>
		</div>
	{/if}

	<!-- Assignment Response -->
	{#if showAssignment}
		<div
			class="absolute top-full left-1/2 transform -translate-x-1/2 mt-2 px-4 py-2
			       bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800
			       rounded-lg shadow-lg z-20 whitespace-nowrap"
		>
			<div class="flex items-center">
				<svg
					class="h-4 w-4 text-green-600 dark:text-green-400 mr-2"
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
				<div class="text-sm">
					<div class="font-medium text-green-900 dark:text-green-100">
						{#if assignedElevator}
							Elevator {assignedElevator}
						{:else}
							Request processed
						{/if}
					</div>
					<div class="text-xs text-green-700 dark:text-green-300">
						{formatFloor(floor)} → {formatFloor(selectedDestination ?? 0)}
					</div>
				</div>
				<button
					class="ml-2 text-green-600 dark:text-green-400 hover:text-green-800 dark:hover:text-green-200"
					on:click={clearAssignment}
					aria-label="Clear assignment"
				>
					✕
				</button>
			</div>
		</div>
	{/if}

	<!-- Floor Selection Modal -->
	<Modal open={showModal} title="Select Destination Floor" size="small" on:close={handleModalClose}>
		<div class="space-y-4">
			<!-- Current Floor Info -->
			<div class="text-center p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
				<div class="text-lg font-semibold text-gray-900 dark:text-white">
					Floor {formatFloor(floor)}
				</div>
				<div class="text-sm text-gray-600 dark:text-gray-400">Select your destination</div>
			</div>

			<!-- Available Destinations Grid -->
			{#if availableDestinations.length > 0}
				<div class="grid grid-cols-4 gap-3">
					{#each availableDestinations as destination}
						<Button
							variant="outline"
							size="medium"
							class="aspect-square hover:scale-105 transition-transform duration-200 
							       min-h-[60px] {getFloorDisplayClass(destination)} 
							       {selectedDestination === destination ? 'ring-2 ring-blue-500' : ''}"
							on:click={() => handleDestinationSelect(destination)}
							disabled={isRequesting || destination === floor}
							ariaLabel="Go to floor {formatFloor(destination)}"
						>
							<div class="text-center">
								<div class="font-bold text-gray-900 dark:text-white">
									{formatFloor(destination)}
								</div>
								{#if destination === 0}
									<div class="text-xs text-gray-600 dark:text-gray-400">Ground</div>
								{:else if destination < 0}
									<div class="text-xs text-gray-600 dark:text-gray-400">Basement</div>
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
				<div class="text-center py-4">
					<div class="spinner mb-2" aria-hidden="true" />
					<p class="text-sm text-gray-600 dark:text-gray-400">Processing request...</p>
				</div>
			{/if}
		</div>
	</Modal>
</div>

<style>
	.floor-call-button {
		/* Metallic effect */
		background: linear-gradient(145deg, #e2e8f0, #cbd5e1);
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.2), inset 0 -1px 0 rgba(0, 0, 0, 0.1),
			0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
	}

	.floor-call-button:hover {
		background: linear-gradient(145deg, #f1f5f9, #e2e8f0);
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3), inset 0 -1px 0 rgba(0, 0, 0, 0.1),
			0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
	}

	.floor-call-button:active {
		transform: scale(0.95);
		box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.1), 0 1px 2px rgba(0, 0, 0, 0.1);
	}

	/* Spinner animation */
	.spinner {
		display: inline-block;
		width: 20px;
		height: 20px;
		border: 2px solid #f3f3f3;
		border-top: 2px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.floor-call-button {
			border: 2px solid currentColor;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.floor-call-button {
			transition: none;
		}

		.floor-call-button:hover {
			transform: none;
		}
	}
</style>
