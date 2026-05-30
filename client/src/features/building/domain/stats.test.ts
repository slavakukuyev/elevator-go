import { describe, it, expect } from 'vitest';
import { computeFleetStats, statusShare } from './stats';
import type { Elevator } from './types';

function makeElevator(overrides: Partial<Elevator> = {}): Elevator {
  return {
    name: 'E',
    minFloor: 0,
    maxFloor: 10,
    currentFloor: 0,
    status: 'idle',
    direction: null,
    requests: 0,
    ...overrides,
  };
}

describe('computeFleetStats', () => {
  it('returns zeroed stats for an empty fleet', () => {
    expect(computeFleetStats([])).toEqual({
      total: 0,
      idle: 0,
      moving: 0,
      deleting: 0,
      fault: 0,
      pendingRequests: 0,
    });
  });

  it('counts elevators by status', () => {
    const stats = computeFleetStats([
      makeElevator({ name: 'A', status: 'idle' }),
      makeElevator({ name: 'B', status: 'moving', requests: 2 }),
      makeElevator({ name: 'C', status: 'moving', requests: 1 }),
      makeElevator({ name: 'D', status: 'deleting' }),
    ]);
    expect(stats.total).toBe(4);
    expect(stats.idle).toBe(1);
    expect(stats.moving).toBe(2);
    expect(stats.deleting).toBe(1);
    expect(stats.fault).toBe(0);
  });

  it('sums pending requests, ignoring negatives', () => {
    const stats = computeFleetStats([
      makeElevator({ name: 'A', requests: 3 }),
      makeElevator({ name: 'B', requests: -1 }),
      makeElevator({ name: 'C', requests: 5 }),
    ]);
    expect(stats.pendingRequests).toBe(8);
  });
});

describe('statusShare', () => {
  it('returns 0 for an empty fleet', () => {
    const stats = computeFleetStats([]);
    expect(statusShare(stats, 'idle')).toBe(0);
  });

  it('returns the percentage of the fleet in a status', () => {
    const stats = computeFleetStats([
      makeElevator({ name: 'A', status: 'idle' }),
      makeElevator({ name: 'B', status: 'idle' }),
      makeElevator({ name: 'C', status: 'moving', requests: 1 }),
      makeElevator({ name: 'D', status: 'moving', requests: 1 }),
    ]);
    expect(statusShare(stats, 'idle')).toBe(50);
    expect(statusShare(stats, 'moving')).toBe(50);
    expect(statusShare(stats, 'fault')).toBe(0);
  });
});
