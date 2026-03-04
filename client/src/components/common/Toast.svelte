<script lang="ts">
	import { fly } from 'svelte/transition';
	import { notifications, removeNotification } from '../../stores/elevators';
	import type { Notification } from '../../types';

	function getIconForType(type: Notification['type']) {
		switch (type) {
			case 'success':
				return {
					path: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
					color: 'text-green-500 dark:text-green-400'
				};
			case 'error':
				return {
					path: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
					color: 'text-red-500 dark:text-red-400'
				};
			case 'warning':
				return {
					path: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
					color: 'text-yellow-500 dark:text-yellow-400'
				};
			case 'info':
			default:
				return {
					path: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
					color: 'text-primary-500 dark:text-primary-400'
				};
		}
	}

	function getBackgroundForType(type: Notification['type']) {
		switch (type) {
			case 'success':
				return 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800';
			case 'error':
				return 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800';
			case 'warning':
				return 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800';
			case 'info':
			default:
				return 'bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700';
		}
	}
</script>

<!-- Toast Container - Upper Right -->
<div class="fixed top-20 right-4 z-[60] space-y-2">
	{#each $notifications as notification (notification.id)}
		<div
			class="{getBackgroundForType(notification.type)} border rounded-lg shadow-lg p-4 min-w-[320px] max-w-md"
			role="alert"
			aria-live={notification.type === 'error' ? 'assertive' : 'polite'}
			transition:fly={{ x: 300, duration: 300 }}
		>
			<div class="flex items-start justify-between">
				<div class="flex items-start">
					<!-- Dynamic Icon -->
					<svg
						class="h-5 w-5 {getIconForType(notification.type).color} mt-0.5 mr-3 flex-shrink-0"
						fill="none"
						viewBox="0 0 24 24"
						stroke="currentColor"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d={getIconForType(notification.type).path}
						/>
					</svg>
					<div class="flex-1">
						<p class="text-sm font-medium text-gray-900 dark:text-gray-100">
							{notification.message}
						</p>
					</div>
				</div>
				<button
					type="button"
					class="ml-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
					aria-label="Close notification"
					on:click={() => removeNotification(notification.id)}
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
