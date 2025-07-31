<script lang="ts">
	import { fly } from 'svelte/transition';
	import { notifications } from '../../stores/elevators';

	function removeNotification(index: number) {
		notifications.update((list) => list.filter((_, i) => i !== index));
	}
</script>

<!-- Toast Container -->
<div class="fixed top-20 right-4 z-40 space-y-2">
	{#each $notifications as notification, index}
		<div
			class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg p-4 min-w-[300px] max-w-md"
			role="alert"
			aria-live="polite"
			transition:fly={{ x: 300, duration: 300 }}
		>
			<div class="flex items-start justify-between">
				<div class="flex items-start">
					<!-- Info Icon -->
					<svg
						class="h-5 w-5 text-primary-500 mt-0.5 mr-3 flex-shrink-0"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
					<div class="flex-1">
						<p class="text-sm font-medium text-gray-900 dark:text-gray-100">
							{notification}
						</p>
					</div>
				</div>
				<button
					type="button"
					class="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
					aria-label="Close notification"
					on:click={() => removeNotification(index)}
				>
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			</div>
		</div>
	{/each}
</div>

<style>
	/* Animation for smooth appearance */
	:global(.toast-enter) {
		transform: translateX(100%);
		opacity: 0;
	}

	:global(.toast-enter-active) {
		transition: all 0.3s ease-out;
	}

	:global(.toast-exit) {
		transform: translateX(100%);
		opacity: 0;
	}

	:global(.toast-exit-active) {
		transition: all 0.3s ease-in;
	}
</style>
