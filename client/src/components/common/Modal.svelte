<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { browser } from '$app/environment';

	export let open = false;
	export let title = '';
	export let size: 'small' | 'medium' | 'large' = 'medium';
	export let closeOnEscape = true;
	export let closeOnClickOutside = true;

	const dispatch = createEventDispatcher();

	let modal: HTMLElement;
	let previouslyFocused: HTMLElement | null = null;

	// Size classes
	$: sizeClasses = {
		small: 'max-w-md',
		medium: 'max-w-lg',
		large: 'max-w-2xl',
	}[size];

	onMount(() => {
		// Focus management
		if (open && browser) {
			previouslyFocused = document.activeElement as HTMLElement;
			modal?.focus();
		}
	});

	$: if (open && browser) {
		// Trap focus within modal
		document.body.style.overflow = 'hidden';
		setTimeout(() => {
			const focusableElements = modal?.querySelectorAll(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			) as NodeListOf<HTMLElement>;

			if (focusableElements.length > 0) {
				focusableElements[0].focus();
			}
		}, 100);
	} else if (browser) {
		document.body.style.overflow = '';
		if (previouslyFocused) {
			previouslyFocused.focus();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (!open || !browser) return;

		if (event.key === 'Escape' && closeOnEscape) {
			close();
		}

		// Trap focus within modal
		if (event.key === 'Tab') {
			const focusableElements = modal?.querySelectorAll(
				'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
			) as NodeListOf<HTMLElement>;

			if (focusableElements.length === 0) return;

			const firstElement = focusableElements[0];
			const lastElement = focusableElements[focusableElements.length - 1];

			if (event.shiftKey) {
				if (document.activeElement === firstElement) {
					lastElement.focus();
					event.preventDefault();
				}
			} else {
				if (document.activeElement === lastElement) {
					firstElement.focus();
					event.preventDefault();
				}
			}
		}
	}

	function handleClickOutside(event: MouseEvent) {
		if (!open || !closeOnClickOutside) return;

		const target = event.target as HTMLElement;
		if (target === event.currentTarget) {
			close();
		}
	}

	function close() {
		open = false;
		dispatch('close');
	}
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 bg-black bg-opacity-50 transition-opacity z-40"
		role="presentation"
		on:click={handleClickOutside}
		on:keydown={handleKeydown}
		transition:fly={{ y: 50, duration: 300 }}
	/>

	<!-- Modal Dialog -->
	<div
		bind:this={modal}
		class="fixed inset-0 flex items-center justify-center p-4 z-50"
		role="presentation"
		on:click={handleClickOutside}
		on:keydown={handleKeydown}
		transition:fly={{ y: 50, duration: 300 }}
	>
		<div
			class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full {sizeClasses} max-h-[90vh] overflow-y-auto"
			role="dialog"
			aria-modal="true"
			aria-labelledby={title ? 'modal-title' : undefined}
			tabindex="-1"
		>
			<!-- Header -->
			{#if title || $$slots.header}
				<div
					class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700"
				>
					<div class="flex items-center">
						{#if title}
							<h3 id="modal-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">
								{title}
							</h3>
						{:else}
							<slot name="header" />
						{/if}
					</div>
					<button
						type="button"
						class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors"
						aria-label="Close modal"
						on:click={close}
					>
						<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>
			{/if}

			<!-- Content -->
			<div class="p-6">
				<slot />
			</div>

			<!-- Footer -->
			{#if $$slots.footer}
				<div
					class="flex items-center justify-end gap-3 p-6 border-t border-gray-200 dark:border-gray-700"
				>
					<slot name="footer" />
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	/* Ensure modal appears above everything */
	:global(.modal-open) {
		overflow: hidden;
	}
</style>
