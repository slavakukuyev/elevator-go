<script lang="ts">
	import { elevators } from '../../stores/elevators';
	import { floorSelectionService } from '../../utils/floorSelection';
	import FloorCallButton from './FloorCallButton.svelte';
	import type { Elevator } from '../../types';

	// Calculate the full range of floors across all elevators
	$: allFloors =
		$elevators.length > 0
			? (() => {
					const minFloor = Math.min(...$elevators.map((e: Elevator) => e.minFloor ?? 0));
					const maxFloor = Math.max(...$elevators.map((e: Elevator) => e.maxFloor ?? 0));
					return Array.from({ length: maxFloor - minFloor + 1 }, (_, i) => maxFloor - i);
			  })()
			: [];

	// Get elevators serving each floor
	function getElevatorsForFloor(floor: number) {
		return $elevators.filter((e: Elevator) => floor >= (e.minFloor ?? 0) && floor <= (e.maxFloor ?? 0));
	}

	// Get elevators at a specific floor
	function getElevatorsAtFloor(floor: number) {
		return $elevators.filter((e: Elevator) => e.currentFloor === floor);
	}

	function formatFloor(floor: number): string {
		return floorSelectionService.formatFloorDisplay(floor);
	}

	function getFloorClass(floor: number): string {
		if (floor === 0)
			return 'bg-yellow-50 dark:bg-yellow-900/10 border-yellow-200 dark:border-yellow-800';
		if (floor < 0) return 'bg-blue-50 dark:bg-blue-900/10 border-blue-200 dark:border-blue-800';
		return 'bg-gray-50 dark:bg-gray-800 border-gray-200 dark:border-gray-700';
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'idle':
				return 'bg-green-500 dark:bg-green-400';
			case 'moving':
				return 'bg-blue-500 dark:bg-blue-400';
			case 'error':
				return 'bg-red-500 dark:bg-red-400';
			default:
				return 'bg-gray-500 dark:bg-gray-400';
		}
	}

	function getDirectionIcon(direction: string | null): string {
		switch (direction) {
			case 'up':
				return '↑';
			case 'down':
				return '↓';
			default:
				return '';
		}
	}

	function handleFloorSelected(event: CustomEvent) {
		const { from, to, assignedElevator } = event.detail;
		console.log(`Floor request: ${from} → ${to}, Assigned: ${assignedElevator}`);
	}
</script>

<div class="building-floor-view h-full bg-gray-50 dark:bg-gray-900 overflow-auto">
	{#if $elevators.length === 0}
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
					Create your first elevator to start using the call button system.
				</p>
				<div
					class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4"
				>
					<h3 class="text-sm font-medium text-blue-900 dark:text-blue-100 mb-2">
						🔘 Call Button System
					</h3>
					<ul class="text-sm text-blue-800 dark:text-blue-200 space-y-1">
						<li>• Click the metallic call button on each floor</li>
						<li>• Select your destination from the modal</li>
						<li>• Get assigned to the optimal elevator</li>
						<li>• Wait for your assigned elevator to arrive</li>
					</ul>
				</div>
			</div>
		</div>
	{:else}
		<!-- Building Floors -->
		<div class="building-floors p-4 space-y-4">
			{#each allFloors as floor (floor)}
				{@const availableElevators = getElevatorsForFloor(floor)}
				{@const elevatorsAtFloor = getElevatorsAtFloor(floor)}

				<div
					class="floor-row {getFloorClass(
						floor
					)} border rounded-lg p-4 transition-all duration-300 hover:shadow-md"
				>
					<div class="flex items-center gap-6 floor-content">
						<!-- Floor Number -->
						<div class="floor-number flex-shrink-0 w-16 text-center">
							<div class="text-xl font-bold text-gray-900 dark:text-white">
								{formatFloor(floor)}
							</div>
							{#if floor === 0}
								<div class="text-xs text-gray-600 dark:text-gray-400">Ground</div>
							{:else if floor < 0}
								<div class="text-xs text-gray-600 dark:text-gray-400">Basement</div>
							{/if}
						</div>

						<!-- Elevator Positions -->
						<div class="elevator-positions flex-shrink-0 min-w-[120px] flex gap-2">
							{#each $elevators as elevator}
								{#if elevator.currentFloor === floor}
									<div
										class="elevator-indicator {getStatusColor(
											elevator.status
										)} rounded-full w-8 h-8 flex items-center justify-center text-white text-sm font-bold"
									>
										{elevator.name ? elevator.name.charAt(0) : '?'}
										{#if elevator.direction}
											<span class="ml-1 text-xs">{getDirectionIcon(elevator.direction)}</span>
										{/if}
									</div>
								{/if}
							{/each}

							{#if elevatorsAtFloor.length === 0}
								<div class="text-xs text-gray-500 dark:text-gray-400 flex items-center h-8">
									No elevators
								</div>
							{/if}
						</div>

						<!-- Floor Call Button -->
						<div class="floor-call-button-container flex-grow flex justify-center">
							{#if availableElevators.length > 0}
								<FloorCallButton {floor} on:floorSelected={handleFloorSelected} />
							{:else}
								<div class="text-center py-4">
									<p class="text-sm text-gray-500 dark:text-gray-400">
										No elevator service available
									</p>
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
