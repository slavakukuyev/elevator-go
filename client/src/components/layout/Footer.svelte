<script lang="ts">
	import { systemStatus } from '../../stores/elevators';
</script>

<footer class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 px-4 sm:px-6 lg:px-8">
	<div class="flex items-center justify-between h-12">
		<!-- Left side - Status -->
		<div class="flex items-center space-x-4">
			<div class="text-sm text-gray-600 dark:text-gray-400">
				System Status: 
				<span class="font-medium {$systemStatus.healthy ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}">
					{$systemStatus.healthy ? 'Healthy' : 'Issues Detected'}
				</span>
			</div>
			<div class="text-sm text-gray-600 dark:text-gray-400">
				Elevators: {$systemStatus.elevatorCount}
			</div>
		</div>

		<!-- Right side - Links -->
		<div class="flex items-center space-x-4">
			<div class="text-sm text-gray-600 dark:text-gray-400">
				© 2024 Elevator Control System
			</div>
			<div class="flex space-x-4">
				<button type="button" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100 transition-colors">
					Documentation
				</button>
				<button type="button" class="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100 transition-colors">
					GitHub
				</button>
			</div>
		</div>
	</div>
</footer>

<style>
	/* Responsive adjustments */
	@media (max-width: 640px) {
		.space-x-4 {
			gap: 0.5rem;
		}
	}
</style> 