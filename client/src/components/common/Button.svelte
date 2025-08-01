<script lang="ts">
	export let variant: 'primary' | 'secondary' | 'danger' | 'success' | 'outline' = 'primary';
	export let size: 'small' | 'medium' | 'large' = 'medium';
	export let disabled = false;
	export let loading = false;
	export let fullWidth = false;
	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let ariaLabel: string | undefined = undefined;
	
	// Allow additional CSS classes
	let className = '';
	export { className as class };

	// Get base classes
	$: baseClasses = [
		'inline-flex items-center justify-center font-medium rounded-md transition-colors duration-200',
		'focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500',
		'disabled:opacity-50 disabled:cursor-not-allowed',
		fullWidth ? 'w-full' : ''
	].join(' ');

	// Get variant classes
	$: variantClasses = {
		primary: 'bg-primary-600 text-white hover:bg-primary-700 dark:bg-primary-500 dark:hover:bg-primary-600',
		secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 dark:bg-gray-700 dark:text-gray-100 dark:hover:bg-gray-600',
		danger: 'bg-red-600 text-white hover:bg-red-700 dark:bg-red-500 dark:hover:bg-red-600',
		success: 'bg-green-600 text-white hover:bg-green-700 dark:bg-green-500 dark:hover:bg-green-600',
		outline: 'border-2 border-primary-600 text-primary-600 hover:bg-primary-50 dark:border-primary-400 dark:text-primary-400 dark:hover:bg-primary-900'
	}[variant];

	// Get size classes
	$: sizeClasses = {
		small: 'px-3 py-1.5 text-sm',
		medium: 'px-4 py-2 text-base',
		large: 'px-6 py-3 text-lg'
	}[size];
</script>

<button
	{type}
	{disabled}
	aria-label={ariaLabel}
	class="{baseClasses} {variantClasses} {sizeClasses} {className}"
	on:click
	on:focus
	on:blur
	on:mouseenter
	on:mouseleave
>
	{#if loading}
		<div class="spinner small mr-2" aria-hidden="true"></div>
		Loading...
	{:else}
		<slot />
	{/if}
</button>

<style>
	button:focus {
		outline: 2px solid var(--primary-500);
		outline-offset: 2px;
	}

	button:disabled {
		pointer-events: none;
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		button {
			border: 2px solid currentColor;
		}
	}
</style> 