// services/websocket.ts - Simplified WebSocket service
import {
    connectionStatus,
    elevators,
    systemStatus,
    addNotification
} from '../stores/elevators';

class ElevatorWebSocketService {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000;
    private reconnectTimer: NodeJS.Timeout | null = null;
    private pingInterval: NodeJS.Timeout | null = null;
    private url: string;

    constructor() {
        this.url = import.meta.env.VITE_WS_URL || 'ws://localhost:6661/ws/status';
    }

    connect() {
        if (this.ws?.readyState === WebSocket.OPEN) {
            console.log('WebSocket already connected');
            return;
        }

        try {
            console.log('Connecting to WebSocket:', this.url);
            this.ws = new WebSocket(this.url);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                connectionStatus.set({
                    connected: true,
                    lastConnected: new Date(),
                    retryCount: 0
                });
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
                this.startPingInterval();
            };

            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.warn('WebSocket closed:', event.code, event.reason);
                connectionStatus.update(status => ({
                    ...status,
                    connected: false
                }));
                this.clearPingInterval();
                this.handleReconnection();
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                connectionStatus.update(status => ({
                    ...status,
                    connected: false
                }));
            };
        } catch (error) {
            console.error('WebSocket connection failed:', error);
            this.handleReconnection();
        }
    }

    private handleMessage(rawData: any) {
        //console.log('📨 Received WebSocket data:', rawData);

        // Backend sends ElevatorStatus format: { "ElevatorName": { name, current_floor, direction, requests, min_floor, max_floor } }
        if (rawData && typeof rawData === 'object' && !Array.isArray(rawData)) {
            const elevatorNames = Object.keys(rawData);

            // Convert backend ElevatorStatus format to client format
            const clientElevators = elevatorNames.map(elevatorName => {
                const backendElevator = rawData[elevatorName];

                // ElevatorStatus uses "requests" field, not "pending_requests"
                const pendingRequests = backendElevator.requests || 0;
                const backendDirection = backendElevator.direction || '';

                // Elevator should be idle when no requests exist, regardless of direction
                const isIdle = pendingRequests === 0;
                const status = isIdle ? 'idle' : 'moving';

                // Convert direction: set to null when idle, otherwise use backend direction
                let direction: 'up' | 'down' | null = null;
                if (!isIdle && backendDirection) {
                    if (backendDirection === 'up' || backendDirection === 'down') {
                        direction = backendDirection as 'up' | 'down';
                    }
                }

                //console.log(`${elevatorName}: requests=${pendingRequests}, direction="${backendDirection}" → status=${status}, clientDirection=${direction}`);

                return {
                    name: backendElevator.name || elevatorName,
                    minFloor: backendElevator.min_floor || 0,
                    maxFloor: backendElevator.max_floor || 10,
                    currentFloor: backendElevator.current_floor || 0,
                    status: status as 'idle' | 'moving' | 'error',
                    direction: direction,
                    doorsOpen: false,
                    hasPassenger: false
                };
            });

            // Update stores
            elevators.set(clientElevators);
            systemStatus.set({
                healthy: true,
                elevatorCount: clientElevators.length,
                lastMaintenance: undefined,
                alerts: []
            });

            // console.log('✅ Updated elevators:', clientElevators);
        } else {
            console.warn('❌ Unexpected WebSocket data format:', rawData);
        }
    }

    private handleReconnection() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
        }

        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts);
            this.reconnectAttempts++;

            connectionStatus.update(status => ({
                ...status,
                retryCount: this.reconnectAttempts
            }));

            console.log(`Reconnecting in ${delay}ms... attempt ${this.reconnectAttempts}`);

            this.reconnectTimer = setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max WebSocket reconnection attempts reached');
            addNotification('Connection lost. Please refresh the page.');
        }
    }

    private startPingInterval() {
        this.pingInterval = setInterval(() => {
            if (this.ws?.readyState === WebSocket.OPEN) {
                this.ws.send(JSON.stringify({ type: 'ping' }));
            }
        }, 30000);
    }

    private clearPingInterval() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    disconnect() {
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        this.clearPingInterval();

        if (this.ws) {
            this.ws.close(1000, 'Client disconnecting');
            this.ws = null;
        }

        connectionStatus.set({
            connected: false,
            retryCount: 0
        });
    }

    send(message: any) {
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.warn('WebSocket not connected. Message not sent:', message);
        }
    }

    isConnected(): boolean {
        return this.ws?.readyState === WebSocket.OPEN;
    }

    getConnectionState(): number {
        return this.ws?.readyState || WebSocket.CLOSED;
    }
}

export const wsService = new ElevatorWebSocketService();

// Auto-connect when the service is imported (only in browser)
if (typeof window !== 'undefined') {
    wsService.connect();

    // Clean up on page unload
    window.addEventListener('beforeunload', () => {
        wsService.disconnect();
    });
} 