<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { elevatorAPI } from '../../services/api';
	import { floorSelectionService } from '../../utils/floorSelection';
	import { elevators } from '../../stores/elevators';
	import type { Elevator } from '../../types';

	export let elevator: Elevator;
	export let fromFloor: number;
	export let isOpen: boolean = false;

	const dispatch = createEventDispatcher<{
		close: void;
		floorSelected: { from: number; to: number; elevatorName: string };
	}>();

	// Generate available floors that can be reached from the current floor (in descending order)
	$: availableFloors = floorSelectionService
		.getAvailableDestinations(fromFloor, $elevators)
		.sort((a, b) => b - a); // Sort in descending order (higher to lower)

	function formatFloor(floor: number): string {
		return floorSelectionService.formatFloorDisplay(floor);
	}

	function getFloorDisplayClass(floor: number): string {
		if (floor === 0)
			return 'bg-yellow-100 dark:bg-yellow-900/20 border-yellow-300 dark:border-yellow-600';
		if (floor < 0) return 'bg-blue-100 dark:bg-blue-900/20 border-blue-300 dark:border-blue-600';
		return 'bg-gray-100 dark:bg-gray-800 border-gray-300 dark:border-gray-600';
	}

	async function handleFloorSelect(toFloor: number) {
		try {
			await elevatorAPI.requestFloor(fromFloor, toFloor);
			dispatch('floorSelected', { from: fromFloor, to: toFloor, elevatorName: elevator.name });
			closePopup();
		} catch (error) {
			console.error('Failed to request floor:', error);
		}
	}

	function closePopup() {
		dispatch('close');
	}

	// Close popup when clicking outside
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closePopup();
		}
	}

	// Close popup when pressing Escape key
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			closePopup();
		}
	}
</script>

{#if isOpen}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 bg-black bg-opacity-50 z-40 flex items-center justify-center"
		on:click={handleBackdropClick}
		on:keydown={handleKeydown}
		role="button"
		tabindex="-1"
		aria-label="Close floor selection popup"
	>
		<!-- Popup Content -->
		<div
			class="popup-container bg-white dark:bg-gray-800 rounded-lg shadow-xl p-2 mx-4 relative"
			role="dialog"
			aria-labelledby="floor-selection-title"
			aria-describedby="floor-selection-description"
		>
			<!-- Close Button -->
			<button
				class="absolute top-1 right-1 w-5 h-5 flex items-center justify-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
				on:click={closePopup}
				aria-label="Close floor selection"
			>
				<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
			</button>

			<!-- Header -->
			<div class="mb-1">
				<h3 id="floor-selection-title" class="text-xs font-semibold text-gray-900 dark:text-white">
					Available Destinations
				</h3>
				<p id="floor-selection-description" class="text-xs text-gray-600 dark:text-gray-400">
					From {formatFloor(fromFloor)} (Reachable floors)
				</p>
			</div>

			<!-- Floor Grid -->
			<div class="grid grid-cols-6 gap-0.5">
				{#each availableFloors as floor}
					<button
						class="floor-button aspect-square p-1 rounded border transition-all duration-200 hover:scale-105 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:ring-offset-1 dark:focus:ring-offset-gray-800 min-w-[40px] max-w-[60px] min-h-[40px] max-h-[60px] {getFloorDisplayClass(
							floor
						)}"
						on:click={() => handleFloorSelect(floor)}
						aria-label="Go to floor {formatFloor(floor)}"
					>
						<span class="text-xs font-bold text-gray-900 dark:text-white">
							{formatFloor(floor)}
						</span>
					</button>
				{/each}
			</div>

			<!-- Cancel Button -->
			<div class="mt-1 flex justify-end">
				<button
					class="px-1 py-0.5 text-xs font-medium text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors"
					on:click={closePopup}
				>
					Cancel
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.floor-button {
		transition: all 0.2s ease-in-out;
		width: clamp(30px, 6vw, 45px);
		height: clamp(30px, 6vw, 45px);
	}

	.floor-button:hover {
		transform: scale(1.05);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.floor-button:active {
		transform: scale(0.98);
	}

	/* Responsive adjustments */
	@media (min-width: 768px) {
		.floor-button {
			width: 35px;
			height: 35px;
		}
	}

	@media (min-width: 1024px) {
		.floor-button {
			width: 32px;
			height: 32px;
		}
	}

	@media (min-width: 1440px) {
		.floor-button {
			width: 30px;
			height: 30px;
		}
	}

	/* Popup container constraints */
	.popup-container {
		max-width: 220px !important;
		width: 220px !important;
	}

	@media (max-width: 480px) {
		.popup-container {
			max-width: 200px !important;
			width: 200px !important;
		}
	}
</style>
