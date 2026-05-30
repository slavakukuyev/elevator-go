// Domain types for the elevator building feature.
// These are the *client* shapes — backend wire shapes live in shared/api and are
// mapped here through a single mapper (see shared/api/mapper.ts).

export type ElevatorStatus = 'idle' | 'moving' | 'deleting' | 'fault';
export type Direction = 'up' | 'down' | null;

export interface Elevator {
  name: string;
  minFloor: number;
  maxFloor: number;
  currentFloor: number;
  status: ElevatorStatus;
  direction: Direction;
  /** Pending request count reported by the backend. */
  requests: number;
}

export interface ElevatorConfig {
  name: string;
  minFloor: number;
  maxFloor: number;
}

export interface FloorRequest {
  from: number;
  to: number;
}

export interface SystemHealth {
  healthy: boolean;
  elevatorCount: number;
}
