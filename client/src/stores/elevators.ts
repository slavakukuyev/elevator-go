// stores/elevators.ts - Svelte stores for elevator state management
import { writable, derived } from 'svelte/store';
import type { Elevator, SystemStatus, ConnectionStatus, FloorRequest, Theme, Notification } from '../types';

// Primary stores
export const elevators = writable<Elevator[]>([]);
export const systemStatus = writable<SystemStatus>({ healthy: true, elevatorCount: 0 });
export const connectionStatus = writable<ConnectionStatus>({
    connected: false,
    retryCount: 0
});
export const currentFloor = writable<number>(0);
export const selectedElevator = writable<Elevator | null>(null);
export const floorRequests = writable<FloorRequest[]>([]);
export const theme = writable<Theme>({ mode: 'light' });

// UI state
export const isLoading = writable(false);
export const showCreateModal = writable(false);
export const showMonitoringPanel = writable(false);
export const showControlPanel = writable(true);
export const notifications = writable<Notification[]>([]);

// Derived stores for performance optimization
export const availableElevators = derived(
    [currentFloor, elevators],
    ([$currentFloor, $elevators]) =>
        $elevators.filter(elev =>
            $currentFloor >= elev.minFloor && $currentFloor <= elev.maxFloor
        )
);

export const idleElevators = derived(
    elevators,
    $elevators => $elevators.filter(elev => elev.status === 'idle')
);

export const movingElevators = derived(
    elevators,
    $elevators => $elevators.filter(elev => elev.status === 'moving')
);

export const errorElevators = derived(
    elevators,
    $elevators => $elevators.filter(elev => elev.status === 'error')
);

export const totalFloors = derived(
    elevators,
    $elevators => {
        if ($elevators.length === 0) return { min: 0, max: 10 };
        const minFloor = Math.min(...$elevators.map(e => e.minFloor));
        const maxFloor = Math.max(...$elevators.map(e => e.maxFloor));
        return { min: minFloor, max: maxFloor };
    }
);

export const pendingRequests = derived(
    floorRequests,
    $requests => $requests.filter(req => req.status === 'pending')
);

export const elevatorUtilization = derived(
    elevators,
    $elevators => {
        const total = $elevators.length;
        const idle = $elevators.filter(e => e.status === 'idle').length;
        const moving = $elevators.filter(e => e.status === 'moving').length;
        const error = $elevators.filter(e => e.status === 'error').length;

        return {
            total,
            idle: total > 0 ? (idle / total) * 100 : 0,
            moving: total > 0 ? (moving / total) * 100 : 0,
            error: total > 0 ? (error / total) * 100 : 0
        };
    }
);

// Helper functions
export function addElevator(elevator: Elevator) {
    elevators.update(list => [...list, elevator]);
    systemStatus.update(status => ({
        ...status,
        elevatorCount: status.elevatorCount + 1
    }));
}

export function updateElevator(name: string, updates: Partial<Elevator>) {
    elevators.update(list =>
        list.map(elevator =>
            elevator.name === name
                ? { ...elevator, ...updates }
                : elevator
        )
    );
}

export function removeElevator(name: string) {
    elevators.update(list => list.filter(elevator => elevator.name !== name));
    systemStatus.update(status => ({
        ...status,
        elevatorCount: Math.max(0, status.elevatorCount - 1)
    }));
}

export function addFloorRequest(request: FloorRequest) {
    floorRequests.update(list => [...list, request]);
}

export function updateFloorRequest(id: string, updates: Partial<FloorRequest>) {
    floorRequests.update(list =>
        list.map(request =>
            request.timestamp.getTime().toString() === id
                ? { ...request, ...updates }
                : request
        )
    );
}

export function addNotification(message: string, type: 'info' | 'success' | 'error' | 'warning' = 'info', duration: number = 5000) {
    const id = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
    const notification: Notification = { id, message, type, duration };

    notifications.update(list => [...list, notification]);

    // Auto-remove after specified duration
    setTimeout(() => {
        notifications.update(list => list.filter(n => n.id !== id));
    }, duration);
}

export function removeNotification(id: string) {
    notifications.update(list => list.filter(n => n.id !== id));
}

export function toggleTheme() {
    theme.update(current => ({
        mode: current.mode === 'light' ? 'dark' : 'light'
    }));
}

export function toggleControlPanel() {
    showControlPanel.update(show => !show);
}

// Initialize with sample data by creating elevators via backend API
export async function initializeSampleData() {
    try {
        isLoading.set(true);
        addNotification('Creating sample elevators...', 'info');

        // Sample elevator configurations
        const sampleConfigs = [
            { name: 'Elevator A', minFloor: -2, maxFloor: 10 },
            { name: 'Elevator B', minFloor: -1, maxFloor: 8 },
            { name: 'Elevator C', minFloor: 0, maxFloor: 12 },
            { name: 'Parking Elevator I', minFloor: -5, maxFloor: 2 },
            { name: 'Parking Elevator J', minFloor: -3, maxFloor: 1 }
        ];

        // Import API service dynamically to avoid circular dependencies
        const { elevatorAPI } = await import('../services/api');

        // Create each elevator via the backend API
        for (const config of sampleConfigs) {
            try {
                await elevatorAPI.createElevator(config);
                // Small delay to avoid overwhelming the backend
                await new Promise(resolve => setTimeout(resolve, 100));
            } catch (error) {
                console.warn(`Failed to create elevator ${config.name}:`, error);
                // Continue with other elevators even if one fails
            }
        }

        addNotification('Sample elevators created successfully!', 'success');

        // WebSocket automatically sends status updates every 100ms
        // No need to explicitly request status

    } catch (error) {
        console.error('Failed to initialize sample data:', error);
        addNotification('Failed to create sample elevators. Please try again.', 'error');
    } finally {
        isLoading.set(false);
    }
} 