import { describe, it, expect } from 'vitest';
import { deriveStatus, deriveDirection, mapElevator, mapBroadcast } from './mapper';
import type { WireElevatorStatus } from './wire';

function wire(overrides: Partial<WireElevatorStatus> = {}): WireElevatorStatus {
  return {
    name: 'A',
    current_floor: 0,
    direction: '',
    requests: 0,
    min_floor: 0,
    max_floor: 10,
    is_deleting: false,
    ...overrides,
  };
}

describe('deriveStatus', () => {
  it('is idle when no requests', () => {
    expect(deriveStatus(wire({ requests: 0, direction: 'up' }))).toBe('idle');
  });

  it('is moving when requests pending', () => {
    expect(deriveStatus(wire({ requests: 2, direction: 'up' }))).toBe('moving');
  });

  it('is deleting when is_deleting flag set', () => {
    expect(deriveStatus(wire({ is_deleting: true, requests: 3 }))).toBe('deleting');
  });

  it('is deleting when direction is "deleting"', () => {
    expect(deriveStatus(wire({ direction: 'deleting' }))).toBe('deleting');
  });
});

describe('deriveDirection', () => {
  it('is null when idle even if backend reports a stale direction', () => {
    expect(deriveDirection(wire({ requests: 0, direction: 'up' }))).toBeNull();
  });

  it('reflects up/down when moving', () => {
    expect(deriveDirection(wire({ requests: 1, direction: 'up' }))).toBe('up');
    expect(deriveDirection(wire({ requests: 1, direction: 'down' }))).toBe('down');
  });

  it('is null for unknown direction strings', () => {
    expect(deriveDirection(wire({ requests: 1, direction: 'deleting' }))).toBeNull();
  });
});

describe('mapElevator', () => {
  it('maps snake_case wire fields to client shape', () => {
    const e = mapElevator('A', wire({ current_floor: 4, min_floor: -2, max_floor: 12, requests: 1, direction: 'up' }));
    expect(e).toEqual({
      name: 'A',
      minFloor: -2,
      maxFloor: 12,
      currentFloor: 4,
      status: 'moving',
      direction: 'up',
      requests: 1,
    });
  });

  it('falls back to the map key when name is missing', () => {
    const e = mapElevator('FromKey', wire({ name: '' }));
    expect(e.name).toBe('FromKey');
  });
});

describe('mapBroadcast', () => {
  it('maps and sorts elevators by name', () => {
    const result = mapBroadcast({
      Charlie: wire({ name: 'Charlie' }),
      Alpha: wire({ name: 'Alpha' }),
      Bravo: wire({ name: 'Bravo' }),
    });
    expect(result.map((e) => e.name)).toEqual(['Alpha', 'Bravo', 'Charlie']);
  });

  it('returns empty array for empty broadcast', () => {
    expect(mapBroadcast({})).toEqual([]);
  });
});
