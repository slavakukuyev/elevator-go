// Pure fleet-statistics helpers. The Svelte client computed these inline inside
// MonitoringDashboard via derived stores; here they are pure functions so the
// summary component stays dumb and the logic is unit-testable.

import type { Elevator, ElevatorStatus } from './types';

export interface FleetStats {
  total: number;
  idle: number;
  moving: number;
  deleting: number;
  fault: number;
  /** Sum of pending requests across the fleet. */
  pendingRequests: number;
}

export function computeFleetStats(elevators: Elevator[]): FleetStats {
  const stats: FleetStats = {
    total: elevators.length,
    idle: 0,
    moving: 0,
    deleting: 0,
    fault: 0,
    pendingRequests: 0,
  };

  for (const e of elevators) {
    stats[e.status] += 1;
    stats.pendingRequests += Math.max(0, e.requests);
  }

  return stats;
}

/** Percentage of the fleet in a given status (0 when the fleet is empty). */
export function statusShare(stats: FleetStats, status: ElevatorStatus): number {
  if (stats.total === 0) return 0;
  return (stats[status] / stats.total) * 100;
}
