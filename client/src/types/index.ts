// types.ts - Type definitions for type safety
export interface Elevator {
    name: string;
    minFloor: number;
    maxFloor: number;
    currentFloor: number;
    status: 'idle' | 'moving' | 'error';
    direction: 'up' | 'down' | null;
    doorsOpen: boolean;
    hasPassenger: boolean;
    threshold?: number;
}

export interface SystemStatus {
    healthy: boolean;
    elevatorCount: number;
    lastMaintenance?: Date;
    alerts?: Alert[];
}

export interface Alert {
    id: string;
    type: 'warning' | 'error' | 'info';
    message: string;
    timestamp: Date;
}

export interface FloorRequest {
    from: number;
    to: number;
    timestamp: Date;
    status: 'pending' | 'assigned' | 'completed' | 'failed';
    elevatorName?: string;
}

export interface WebSocketMessage {
    type: 'status' | 'elevator_update' | 'floor_request' | 'system_alert';
    payload: any;
    timestamp: Date;
}

export interface ElevatorConfig {
    name: string;
    minFloor: number;
    maxFloor: number;
    threshold?: number;
    capacity?: number;
}

export interface Theme {
    mode: 'light' | 'dark';
}

export interface ConnectionStatus {
    connected: boolean;
    lastConnected?: Date;
    retryCount: number;
}

export interface MetricsData {
    totalRequests: number;
    averageWaitTime: number;
    elevatorUtilization: Record<string, number>;
    peakHours: number[];
}

export interface ValidationResult {
    valid: boolean;
    message?: string;
    elevator?: Elevator;
}

export interface FloorSelectionData {
    currentFloor: number;
    availableFloors: number[];
    selectedFloor?: number;
}

// API Error Response Types
export interface APIErrorResponse {
    success: false;
    error: {
        code: string;
        message: string;
        details?: string;
        request_id: string;
        user_message?: string;
    };
    meta: {
        request_id: string;
        version: string;
        duration: string;
    };
    timestamp: string;
}

export interface APIError {
    code: string;
    message: string;
    details?: string;
    userMessage?: string;
    requestId: string;
    timestamp: string;
    rawError?: string;
} 