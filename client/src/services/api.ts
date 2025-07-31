// services/api.ts - HTTP API service with TypeScript
import type { Elevator, ElevatorConfig, SystemStatus, FloorRequest, MetricsData } from '../types';
import { addElevator, addNotification } from '../stores/elevators';

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:6660/v1';

// Backend response types - V1 API format with wrapped responses
interface BackendV1Response<T> {
    success: boolean;
    data: T;
    meta: {
        request_id: string;
        version: string;
        duration: string;
    };
    timestamp: string;
}

interface BackendElevatorCreateData {
    name: string;
    min_floor: number;
    max_floor: number;
    message: string;
}

type BackendElevatorCreateResponse = BackendV1Response<BackendElevatorCreateData>;

interface BackendFloorRequestData {
    elevator_name: string;
    from_floor: number;
    to_floor: number;
    direction: string;
    message: string;
}

type BackendFloorRequestResponse = BackendV1Response<BackendFloorRequestData>;

interface BackendHealthResponseData {
    status: string;
    timestamp: string;
    checks: {
        elevators: Record<string, any>;
        healthy_elevators: number;
        elevators_count: number;
        active_requests: number;
        system_healthy: boolean;
        timestamp: string;
        total_elevators: number;
    };
}

type BackendHealthResponse = BackendV1Response<BackendHealthResponseData>;

interface BackendMetricsResponseData {
    timestamp: string;
    metrics: Record<string, any>;
}

type BackendMetricsResponse = BackendV1Response<BackendMetricsResponseData>;

class APIService {
    private async request<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<T> {
        const url = `${API_BASE}${endpoint}`;
        const config: RequestInit = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`API Error ${response.status}: ${errorText}`);
            }

            // Handle empty responses
            const text = await response.text();
            return text ? JSON.parse(text) : ({} as T);
        } catch (error) {
            console.error(`API request failed: ${endpoint}`, error);
            throw error;
        }
    }

    // Transform backend elevator response to client format
    private transformElevatorResponse(backendResponse: BackendElevatorCreateResponse): Elevator {
        const data = backendResponse.data;
        return {
            name: data.name,
            minFloor: data.min_floor,
            maxFloor: data.max_floor,
            currentFloor: data.min_floor, // Default to min floor
            status: 'idle',
            direction: null,
            doorsOpen: false,
            hasPassenger: false
        };
    }

    // Transform backend floor request response to client format
    private transformFloorRequestResponse(backendResponse: BackendFloorRequestResponse): FloorRequest {
        const data = backendResponse.data;
        return {
            from: data.from_floor,
            to: data.to_floor,
            timestamp: new Date(),
            status: 'assigned',
            elevatorName: data.elevator_name
        };
    }

    // Transform backend health response to client format
    private transformHealthResponse(backendResponse: BackendHealthResponse): SystemStatus {
        return {
            healthy: backendResponse.data.status === 'healthy',
            elevatorCount: backendResponse.data.checks.total_elevators || 0,
            lastMaintenance: undefined,
            alerts: []
        };
    }

    // Transform backend metrics response to client format
    private transformMetricsResponse(backendResponse: BackendMetricsResponse): MetricsData {
        return {
            totalRequests: backendResponse.data.metrics.total_requests || 0,
            averageWaitTime: backendResponse.data.metrics.average_response_time || 0,
            elevatorUtilization: {},
            peakHours: []
        };
    }

    async createElevator(config: ElevatorConfig): Promise<Elevator> {
        const response = await this.request<BackendElevatorCreateResponse>('/elevators', {
            method: 'POST',
            body: JSON.stringify({
                name: config.name,
                min_floor: config.minFloor,
                max_floor: config.maxFloor
                // Note: capacity is not supported by backend
            })
        });

        const elevator = this.transformElevatorResponse(response);

        if (elevator) {
            addElevator(elevator);
            addNotification(`Elevator ${elevator.name} created successfully`);
        }

        return elevator;
    }

    async requestFloor(fromFloor: number, toFloor: number): Promise<FloorRequest> {
        const response = await this.request<BackendFloorRequestResponse>('/floors/request', {
            method: 'POST',
            body: JSON.stringify({ from: fromFloor, to: toFloor })
        });

        const floorRequest = this.transformFloorRequestResponse(response);

        if (floorRequest) {
            addNotification(`Floor request from ${fromFloor} to ${toFloor} submitted`);
        }

        return floorRequest;
    }

    async getHealthStatus(): Promise<SystemStatus | null> {
        try {
            const response = await this.request<BackendHealthResponse>('/health');
            console.log('Health check response:', response);
            const transformed = this.transformHealthResponse(response);
            console.log('Transformed health status:', transformed);
            return transformed;
        } catch (error) {
            console.error('Health check failed:', error);
            return null;
        }
    }

    // Get current elevator status from backend and sync with stores
    async syncElevatorStatus(): Promise<void> {
        try {
            const response = await this.request<BackendHealthResponse>('/health');
            if (response.data.checks.elevators) {
                const backendElevators = Object.values(response.data.checks.elevators).map((elevator: any) => {
                    // Health endpoint uses "pending_requests", WebSocket uses "requests"
                    const pendingRequests = elevator.pending_requests || 0;
                    const backendDirection = elevator.direction || '';

                    let status: 'idle' | 'moving' | 'error' = 'idle';
                    let direction: 'up' | 'down' | null = null;

                    if (!elevator.is_healthy) {
                        status = 'error';
                    } else {
                        // Elevator should be idle when no requests exist, regardless of direction
                        const isIdle = pendingRequests === 0;
                        status = isIdle ? 'idle' : 'moving';

                        // Convert direction: set to null when idle, otherwise use backend direction
                        if (!isIdle && backendDirection) {
                            if (backendDirection === 'up' || backendDirection === 'down') {
                                direction = backendDirection as 'up' | 'down';
                            }
                        }
                    }

                    return {
                        name: elevator.name,
                        minFloor: elevator.min_floor,
                        maxFloor: elevator.max_floor,
                        currentFloor: elevator.current_floor,
                        status: status,
                        direction: direction,
                        doorsOpen: false, // Backend doesn't provide this info
                        hasPassenger: false // Backend doesn't provide this info
                    };
                });

                // Import and update the elevators store
                const { elevators, systemStatus } = await import('../stores/elevators');
                elevators.set(backendElevators);
                systemStatus.update(status => ({
                    ...status,
                    elevatorCount: backendElevators.length,
                    healthy: response.data.checks.system_healthy
                }));

                console.log('Synced elevator status from backend:', backendElevators);
            }
        } catch (error) {
            console.error('Failed to sync elevator status:', error);
        }
    }

    async getMetrics(): Promise<MetricsData | null> {
        try {
            const response = await this.request<BackendMetricsResponse>('/v1/metrics');
            return this.transformMetricsResponse(response);
        } catch (error) {
            console.error('Metrics fetch failed:', error);
            return null;
        }
    }

    // Note: The following methods are not implemented in the backend API
    // They are kept for future implementation or can be removed if not needed

    async getElevators(): Promise<Elevator[]> {
        console.warn('getElevators: Not implemented in backend API');
        return [];
    }

    async getElevator(_name: string): Promise<Elevator> {
        console.warn('getElevator: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async deleteElevator(_name: string): Promise<void> {
        console.warn('deleteElevator: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async getFloorRequests(): Promise<FloorRequest[]> {
        console.warn('getFloorRequests: Not implemented in backend API');
        return [];
    }

    async callElevator(_floor: number, _direction: 'up' | 'down'): Promise<void> {
        console.warn('callElevator: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async getStatus(): Promise<{ elevators: Record<string, Elevator>; system: SystemStatus }> {
        console.warn('getStatus: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async emergencyStop(_elevatorName: string): Promise<void> {
        console.warn('emergencyStop: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async emergencyRelease(_elevatorName: string): Promise<void> {
        console.warn('emergencyRelease: Not implemented in backend API');
        throw new Error('Not implemented');
    }

    async scheduleMaintenance(_elevatorName: string, _dateTime: Date): Promise<void> {
        console.warn('scheduleMaintenance: Not implemented in backend API');
        throw new Error('Not implemented');
    }
}

export const elevatorAPI = new APIService(); 