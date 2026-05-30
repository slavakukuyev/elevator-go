import { describe, it, expect } from 'vitest';
import {
  formatFloor,
  buildingFloorsDescending,
  reachableDestinations,
  elevatorsServingFloor,
} from './floors';
import type { Elevator } from './types';

function elevator(overrides: Partial<Elevator> = {}): Elevator {
  return {
    name: 'A',
    minFloor: 0,
    maxFloor: 10,
    currentFloor: 0,
    status: 'idle',
    direction: null,
    requests: 0,
    ...overrides,
  };
}

describe('formatFloor', () => {
  it('formats ground, basement, and upper floors', () => {
    expect(formatFloor(0)).toBe('G');
    expect(formatFloor(-3)).toBe('B3');
    expect(formatFloor(7)).toBe('7');
    expect(formatFloor(undefined)).toBe('?');
  });
});

describe('buildingFloorsDescending', () => {
  it('spans the union of all elevator ranges, top-first', () => {
    const result = buildingFloorsDescending([
      elevator({ minFloor: -2, maxFloor: 5 }),
      elevator({ minFloor: 0, maxFloor: 10 }),
    ]);
    expect(result[0]).toBe(10);
    expect(result[result.length - 1]).toBe(-2);
  });

  it('is empty with no elevators', () => {
    expect(buildingFloorsDescending([])).toEqual([]);
  });
});

describe('reachableDestinations', () => {
  it('returns in-range floors excluding the origin', () => {
    const result = reachableDestinations(0, [elevator({ minFloor: -1, maxFloor: 3 })]);
    expect(result).toEqual([-1, 1, 2, 3]);
  });

  it('excludes elevators that do not serve the origin floor', () => {
    const result = reachableDestinations(8, [elevator({ minFloor: 0, maxFloor: 5 })]);
    expect(result).toEqual([]);
  });

  it('ignores faulted and deleting elevators', () => {
    const result = reachableDestinations(0, [
      elevator({ status: 'deleting', minFloor: 0, maxFloor: 9 }),
      elevator({ status: 'fault', minFloor: 0, maxFloor: 9 }),
    ]);
    expect(result).toEqual([]);
  });

  it('unions destinations across multiple serving elevators', () => {
    const result = reachableDestinations(0, [
      elevator({ name: 'A', minFloor: 0, maxFloor: 2 }),
      elevator({ name: 'B', minFloor: -2, maxFloor: 0 }),
    ]);
    expect(result).toEqual([-2, -1, 1, 2]);
  });
});

describe('elevatorsServingFloor', () => {
  it('filters elevators by range', () => {
    const a = elevator({ name: 'A', minFloor: 0, maxFloor: 5 });
    const b = elevator({ name: 'B', minFloor: 6, maxFloor: 10 });
    expect(elevatorsServingFloor(3, [a, b])).toEqual([a]);
  });
});
