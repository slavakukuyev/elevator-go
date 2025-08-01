<script lang="ts">
	import type { Elevator } from '../../types';
	import { floorSelectionService } from '../../utils/floorSelection';

	export let elevator: Elevator;

	$: floorRange =
		elevator.minFloor !== undefined && elevator.maxFloor !== undefined
			? Array.from(
					{ length: elevator.maxFloor - elevator.minFloor + 1 },
					(_, i) => elevator.maxFloor - i
			  )
			: [];

	// Track previous position for smooth animations
	let previousFloor = elevator.currentFloor;

	$: elevatorPosition = calculateElevatorPosition(
		elevator.currentFloor,
		elevator.minFloor,
		elevator.maxFloor
	);

	// Calculate total height based on number of floors
	$: totalHeight = floorRange.length * 60; // 60px per floor

	// Update previous floor when current floor changes
	$: if (elevator.currentFloor !== previousFloor) {
		previousFloor = elevator.currentFloor;
	}

	function calculateElevatorPosition(
		currentFloor: number,
		minFloor: number,
		maxFloor: number
	): number {
		if (minFloor === undefined || maxFloor === undefined || currentFloor === undefined) {
			return 0;
		}
		const totalFloors = maxFloor - minFloor + 1;
		const floorIndex = currentFloor - minFloor;
		return (totalFloors - floorIndex - 1) * 60; // 60px per floor
	}

	function formatFloor(floor: number | undefined): string {
		return floorSelectionService.formatFloorDisplay(floor);
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
</script>

<div
	class="bg-white dark:bg-gray-800 rounded-lg shadow-elevator hover:shadow-elevator-hover transition-shadow duration-200 p-4"
	role="region"
	aria-labelledby="elevator-{elevator.name}-title"
>
	<!-- Elevator Header -->
	<div class="mb-4">
		<h3
			id="elevator-{elevator.name}-title"
			class="text-lg font-semibold text-gray-900 dark:text-white"
		>
			{elevator.name}
		</h3>
		<div class="flex items-center justify-between text-sm text-gray-600 dark:text-gray-400 mt-1">
			<span>Floors {elevator.minFloor ?? '?'} - {elevator.maxFloor ?? '?'}</span>
			<div class="flex items-center space-x-2">
				<div class="w-2 h-2 rounded-full {getStatusColor(elevator.status)}" />
				<span class="capitalize">{elevator.status}</span>
				{#if elevator.direction}
					<span class="font-bold">{getDirectionIcon(elevator.direction)}</span>
				{/if}
			</div>
		</div>
	</div>

	<!-- Elevator Shaft -->
	<div
		class="elevator-shaft relative bg-gray-100 dark:bg-gray-700 rounded-lg p-2 overflow-visible"
		style="min-height: {totalHeight + 60}px;"
	>
		<!-- Floor Indicators -->
		<div class="absolute left-0 top-0 w-12" style="height: {totalHeight}px;">
			{#each floorRange as floor}
				<div
					class="floor-indicator absolute w-full flex items-center justify-center h-[60px] text-xs font-medium {floor ===
					0
						? 'text-yellow-600 dark:text-yellow-400 font-bold'
						: floor === elevator.currentFloor
						? 'text-green-600 dark:text-green-400 font-bold'
						: 'text-gray-600 dark:text-gray-400'}"
					style="top: {calculateElevatorPosition(floor, elevator.minFloor, elevator.maxFloor)}px;"
				>
					{formatFloor(floor)}
					{#if floor === 0}
						<div
							class="absolute -bottom-1 left-1/2 transform -translate-x-1/2 w-2 h-0.5 bg-yellow-500 dark:bg-yellow-400"
						/>
					{:else if floor === elevator.currentFloor}
						<div
							class="absolute -bottom-1 left-1/2 transform -translate-x-1/2 w-2 h-0.5 bg-green-500 dark:bg-green-400"
						/>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Elevator Shaft Visual -->
		<div
			class="ml-16 mr-16 relative border-2 border-gray-300 dark:border-gray-600 rounded"
			style="height: {totalHeight}px;"
		>
			<!-- Zero Floor Line -->
			{#if elevator.minFloor <= 0 && elevator.maxFloor >= 0}
				<div
					class="absolute left-0 right-0 h-0.5 bg-yellow-500 dark:bg-yellow-400 z-10"
					style="top: {calculateElevatorPosition(0, elevator.minFloor, elevator.maxFloor) + 30}px;"
				/>
			{/if}

			<!-- Floor Lines -->
			{#each floorRange as floor}
				<div
					class="absolute left-0 right-0 h-px bg-gray-300 dark:bg-gray-600"
					style="top: {calculateElevatorPosition(floor, elevator.minFloor, elevator.maxFloor) +
						30}px;"
				/>
			{/each}

			<!-- Elevator Car -->
			<div
				class="elevator-car absolute left-1 right-1 h-[60px] rounded-lg {getStatusColor(
					elevator.status
				)} flex items-center justify-center text-white font-bold shadow-lg z-20 {elevator.status ===
				'moving'
					? 'animate-pulse'
					: ''}"
				style="top: {elevatorPosition + 2}px;"
				data-testid="elevator-car"
				data-current-floor={elevator.currentFloor}
				data-status={elevator.status}
			>
				<div class="text-center">
					<div class="text-sm font-bold">{formatFloor(elevator.currentFloor)}</div>
					{#if elevator.direction}
						<div class="text-lg animate-bounce">{getDirectionIcon(elevator.direction)}</div>
					{/if}
					{#if elevator.status === 'moving'}
						<div class="text-xs opacity-75 mt-1">Moving</div>
					{/if}
				</div>

				<!-- Doors -->
				<div class="doors {elevator.doorsOpen ? 'open' : ''} absolute inset-0 rounded-lg">
					<div class="door-left" />
					<div class="door-right" />
				</div>
			</div>
		</div>
	</div>

	<!-- Status Info -->
	<div class="mt-4 text-sm text-gray-600 dark:text-gray-400">
		<div class="flex justify-between">
			<span>Current Floor:</span>
			<span class="font-medium text-gray-900 dark:text-white"
				>{formatFloor(elevator.currentFloor)}</span
			>
		</div>
		{#if elevator.hasPassenger}
			<div class="flex justify-between">
				<span>Passenger:</span>
				<span class="font-medium text-primary-600 dark:text-primary-400">Yes</span>
			</div>
		{/if}
		<div class="flex justify-between">
			<span>Doors:</span>
			<span
				class="font-medium {elevator.doorsOpen
					? 'text-green-600 dark:text-green-400'
					: 'text-gray-600 dark:text-gray-400'}"
			>
				{elevator.doorsOpen ? 'Open' : 'Closed'}
			</span>
		</div>
	</div>
</div>

<style>
	/* Elevator car animation */
	.elevator-car {
		transition: top 0.8s ease-in-out;
	}

	/* Door animations */
	.doors {
		position: relative;
		overflow: hidden;
	}

	.door-left,
	.door-right {
		position: absolute;
		top: 0;
		bottom: 0;
		width: 50%;
		background: rgba(0, 0, 0, 0.1);
		transition: transform 0.5s ease-in-out;
	}

	.door-left {
		left: 0;
		transform: translateX(0);
	}

	.door-right {
		right: 0;
		transform: translateX(0);
	}

	.doors.open .door-left {
		transform: translateX(-100%);
	}

	.doors.open .door-right {
		transform: translateX(100%);
	}

	/* Floor indicator animations */
	.floor-indicator {
		transition: color 0.3s ease-in-out;
	}
</style>
