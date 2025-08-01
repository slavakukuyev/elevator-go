<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Modal from '../common/Modal.svelte';
	import Button from '../common/Button.svelte';
	import { elevatorAPI } from '../../services/api';
	import { elevators } from '../../stores/elevators';
	import { validationService } from '../../utils/validation';
	import type { ElevatorConfig, APIError } from '../../types';
	import type { ValidationError } from '../../utils/validation';

	export let open = false;

	const dispatch = createEventDispatcher();

	function generateDefaultName(): string {
		const existingNames = $elevators.map((e) => e.name);
		let counter = 1;
		let defaultName = `Elevator-${String.fromCharCode(64 + counter)}`; // Start with Elevator-A

		while (existingNames.includes(defaultName)) {
			counter++;
			defaultName = `Elevator-${String.fromCharCode(64 + counter)}`;
		}

		return defaultName;
	}

	let formData: ElevatorConfig = {
		name: generateDefaultName(),
		minFloor: 0,
		maxFloor: 10,
		threshold: 5,
	};

	let errors: ValidationError[] = [];
	let apiError: APIError | null = null;
	let isSubmitting = false;

	function resetForm() {
		formData = {
			name: generateDefaultName(),
			minFloor: 0,
			maxFloor: 10,
			threshold: 5,
		};
		errors = [];
		apiError = null;
		isSubmitting = false;
	}

	function handleClose() {
		resetForm();
		open = false;
		dispatch('close');
	}

	function parseAPIError(error: Error): APIError {
		try {
			// Extract JSON from error message if it contains API Error
			const errorMessage = error.message;
			const match = errorMessage.match(/API Error \d+: (.+)/);

			if (match) {
				const jsonStr = match[1];
				const apiResponse = JSON.parse(jsonStr);

				// Check if it's a structured API error response
				if (apiResponse.success === false && apiResponse.error) {
					return {
						code: apiResponse.error.code || 'UNKNOWN_ERROR',
						message: apiResponse.error.message || 'An error occurred',
						details: apiResponse.error.details,
						userMessage: apiResponse.error.user_message,
						requestId: apiResponse.error.request_id || apiResponse.meta?.request_id,
						timestamp: apiResponse.timestamp || new Date().toISOString(),
						rawError: errorMessage,
					};
				}
			}

			// Fallback for non-structured errors
			return {
				code: 'UNKNOWN_ERROR',
				message: 'Failed to create elevator',
				details: errorMessage,
				userMessage: 'Please check your input and try again.',
				requestId: 'unknown',
				timestamp: new Date().toISOString(),
				rawError: errorMessage,
			};
		} catch (parseError) {
			// If parsing fails, return a generic error
			return {
				code: 'PARSE_ERROR',
				message: 'Failed to create elevator',
				details: error.message,
				userMessage: 'Please check your input and try again.',
				requestId: 'unknown',
				timestamp: new Date().toISOString(),
				rawError: error.message,
			};
		}
	}

	async function handleSubmit() {
		if (isSubmitting) return;

		// Clear previous errors
		errors = [];
		apiError = null;

		// Validate form
		const configValidation = validationService.validateElevatorConfig(formData);
		const uniqueNameValidation = validationService.validateElevatorNameUniqueness(
			formData.name,
			$elevators.map((e) => e.name)
		);

		errors = [...configValidation.errors, ...uniqueNameValidation.errors];

		if (errors.length > 0) {
			return;
		}

		isSubmitting = true;

		try {
			await elevatorAPI.createElevator(formData);
			handleClose();
		} catch (error) {
			console.error('Failed to create elevator:', error);
			apiError = parseAPIError(error as Error);
		} finally {
			isSubmitting = false;
		}
	}

	function getFieldError(fieldName: string): string | null {
		return validationService.getFieldError(errors, fieldName);
	}

	function hasFieldError(fieldName: string): boolean {
		return validationService.hasFieldError(errors, fieldName);
	}
</script>

<Modal bind:open title="Create New Elevator" size="medium" on:close={handleClose}>
	<form on:submit|preventDefault={handleSubmit} class="space-y-4">
		<!-- API Error Display -->
		{#if apiError}
			<div
				class="bg-red-50 border border-red-200 rounded-md p-4 dark:bg-red-900/20 dark:border-red-800"
			>
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
								clip-rule="evenodd"
							/>
						</svg>
					</div>
					<div class="ml-3 flex-1">
						<h3 class="text-sm font-medium text-red-800 dark:text-red-200">
							{apiError.userMessage || apiError.message}
						</h3>

						{#if apiError.details}
							<div class="mt-2 text-sm text-red-700 dark:text-red-300">
								<p class="font-medium">Details:</p>
								<p class="mt-1">{apiError.details}</p>
							</div>
						{/if}

						<div class="mt-3 text-xs text-red-600 dark:text-red-400">
							<div class="flex flex-wrap gap-4">
								<span
									>Error Code: <span
										class="font-mono bg-red-100 dark:bg-red-800 px-1 py-0.5 rounded"
										>{apiError.code}</span
									></span
								>
								{#if apiError.requestId !== 'unknown'}
									<span
										>Request ID: <span
											class="font-mono bg-red-100 dark:bg-red-800 px-1 py-0.5 rounded"
											>{apiError.requestId}</span
										></span
									>
								{/if}
								<span
									>Time: <span class="font-mono bg-red-100 dark:bg-red-800 px-1 py-0.5 rounded"
										>{new Date(apiError.timestamp).toLocaleTimeString()}</span
									></span
								>
							</div>
						</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- General Validation Error -->
		{#if getFieldError('general')}
			<div
				class="bg-red-50 border border-red-200 rounded-md p-3 dark:bg-red-900/20 dark:border-red-800"
			>
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
								clip-rule="evenodd"
							/>
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800 dark:text-red-200">
							{getFieldError('general')}
						</h3>
					</div>
				</div>
			</div>
		{/if}

		<!-- Name Field -->
		<div>
			<label for="elevator-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
				Elevator Name <span class="text-red-500">*</span>
			</label>
			<input
				id="elevator-name"
				type="text"
				bind:value={formData.name}
				class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white {hasFieldError(
					'name'
				)
					? 'border-red-500 focus:border-red-500 focus:ring-red-500'
					: ''}"
				placeholder="e.g., Main-A, Tower-1, Service"
				required
			/>
			{#if getFieldError('name')}
				<p class="mt-2 text-sm text-red-600 dark:text-red-400">{getFieldError('name')}</p>
			{/if}
		</div>

		<!-- Floor Range -->
		<div class="grid grid-cols-2 gap-4">
			<div>
				<label for="min-floor" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
					Min Floor <span class="text-red-500">*</span>
				</label>
				<input
					id="min-floor"
					type="number"
					bind:value={formData.minFloor}
					class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white {hasFieldError(
						'minFloor'
					)
						? 'border-red-500 focus:border-red-500 focus:ring-red-500'
						: ''}"
					min="-10"
					max="100"
					required
				/>
				{#if getFieldError('minFloor')}
					<p class="mt-2 text-sm text-red-600 dark:text-red-400">{getFieldError('minFloor')}</p>
				{/if}
			</div>
			<div>
				<label for="max-floor" class="block text-sm font-medium text-gray-700 dark:text-gray-300">
					Max Floor <span class="text-red-500">*</span>
				</label>
				<input
					id="max-floor"
					type="number"
					bind:value={formData.maxFloor}
					class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white {hasFieldError(
						'maxFloor'
					)
						? 'border-red-500 focus:border-red-500 focus:ring-red-500'
						: ''}"
					min="-10"
					max="100"
					required
				/>
				{#if getFieldError('maxFloor')}
					<p class="mt-2 text-sm text-red-600 dark:text-red-400">{getFieldError('maxFloor')}</p>
				{/if}
			</div>
		</div>

		<!-- Threshold Field -->
		<div>
			<label
				for="elevator-threshold"
				class="block text-sm font-medium text-gray-700 dark:text-gray-300"
			>
				Threshold
			</label>
			<input
				id="elevator-threshold"
				type="number"
				bind:value={formData.threshold}
				class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white {hasFieldError(
					'threshold'
				)
					? 'border-red-500 focus:border-red-500 focus:ring-red-500'
					: ''}"
				min="1"
				max="100"
				placeholder="5"
			/>
			<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
				Maximum number of requests the elevator can handle before going into maintenance mode
			</p>
			{#if getFieldError('threshold')}
				<p class="mt-2 text-sm text-red-600 dark:text-red-400">{getFieldError('threshold')}</p>
			{/if}
		</div>

		<!-- Floor Range Preview -->
		<div class="bg-gray-50 dark:bg-gray-700 rounded-md p-3">
			<h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Preview</h4>
			<div class="text-sm text-gray-600 dark:text-gray-400">
				{#if formData.minFloor !== undefined && formData.maxFloor !== undefined}
					<p>
						<strong>Floors served:</strong>
						{formData.minFloor} to {formData.maxFloor}
						({formData.maxFloor - formData.minFloor + 1} floors)
					</p>
				{/if}
				{#if formData.threshold}
					<p>
						<strong>Threshold:</strong>
						{formData.threshold} requests
					</p>
				{/if}
			</div>
		</div>
	</form>

	<svelte:fragment slot="footer">
		<Button variant="secondary" on:click={handleClose}>Cancel</Button>
		<Button
			type="submit"
			variant="primary"
			disabled={isSubmitting}
			loading={isSubmitting}
			on:click={handleSubmit}
		>
			{isSubmitting ? 'Creating...' : 'Create Elevator'}
		</Button>
	</svelte:fragment>
</Modal>

<style>
	/* Form styling */
	input:focus {
		outline: none;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	input[type='number']::-webkit-outer-spin-button,
	input[type='number']::-webkit-inner-spin-button {
		-webkit-appearance: none;
		margin: 0;
	}

	input[type='number'] {
		-moz-appearance: textfield;
	}
</style>
