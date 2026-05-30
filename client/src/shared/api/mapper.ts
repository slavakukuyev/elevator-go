// Single backend→client mapper. The old Svelte client duplicated this logic in
// both api.ts and websocket.ts and they drifted (one used `requests`, the other
// `pending_requests`). Here it lives once. (north-star design §4.4.)

import type { Elevator, ElevatorStatus, Direction } from '@/features/building/domain/types';
import type { WireElevatorStatus } from './wire';

/**
 * Derive client status from the backend's raw fields.
 *
 * Backend semantics (internal/domain/direction.go):
 *   - direction "deleting" or is_deleting → 'deleting'
 *   - requests === 0 → 'idle' (regardless of stale direction)
 *   - otherwise → 'moving'
 *
 * 'fault' is reserved for a future health signal; the status broadcast never
 * reports faults today, so it is not produced here.
 */
export function deriveStatus(wire: WireElevatorStatus): ElevatorStatus {
  if (wire.is_deleting || wire.direction === 'deleting') return 'deleting';
  return wire.requests > 0 ? 'moving' : 'idle';
}

export function deriveDirection(wire: WireElevatorStatus): Direction {
  // Idle / deleting elevators show no directional arrow.
  if (wire.requests <= 0) return null;
  if (wire.direction === 'up' || wire.direction === 'down') return wire.direction;
  return null;
}

export function mapElevator(name: string, wire: WireElevatorStatus): Elevator {
  return {
    name: wire.name || name,
    minFloor: wire.min_floor,
    maxFloor: wire.max_floor,
    currentFloor: wire.current_floor,
    status: deriveStatus(wire),
    direction: deriveDirection(wire),
    requests: wire.requests ?? 0,
  };
}

/** Map a full WS broadcast (keyed by name) into a sorted Elevator[]. */
export function mapBroadcast(raw: Record<string, WireElevatorStatus>): Elevator[] {
  return Object.entries(raw)
    .map(([name, wire]) => mapElevator(name, wire))
    .sort((a, b) => a.name.localeCompare(b.name));
}
