import { describe, it, expect, beforeEach } from 'vitest';
import { useLiveStore } from './liveStore';
import type { Elevator } from '@/features/building/domain/types';

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

describe('liveStore.setSnapshot', () => {
  beforeEach(() => useLiveStore.getState().reset());

  it('stores elevators keyed by name with order', () => {
    useLiveStore.getState().setSnapshot([elevator({ name: 'A' }), elevator({ name: 'B' })]);
    const { byName, order } = useLiveStore.getState();
    expect(order).toEqual(['A', 'B']);
    expect(byName.A.name).toBe('A');
  });

  it('preserves object identity for unchanged elevators', () => {
    const store = useLiveStore.getState();
    store.setSnapshot([elevator({ name: 'A', currentFloor: 3 })]);
    const first = useLiveStore.getState().byName.A;

    // Same data again → identity must be preserved so memoized rows skip render.
    store.setSnapshot([elevator({ name: 'A', currentFloor: 3 })]);
    expect(useLiveStore.getState().byName.A).toBe(first);
  });

  it('replaces identity when an elevator changes', () => {
    const store = useLiveStore.getState();
    store.setSnapshot([elevator({ name: 'A', currentFloor: 3 })]);
    const first = useLiveStore.getState().byName.A;

    store.setSnapshot([elevator({ name: 'A', currentFloor: 4 })]);
    expect(useLiveStore.getState().byName.A).not.toBe(first);
    expect(useLiveStore.getState().byName.A.currentFloor).toBe(4);
  });

  it('drops elevators no longer in the snapshot (completed deletion)', () => {
    const store = useLiveStore.getState();
    store.setSnapshot([elevator({ name: 'A' }), elevator({ name: 'B' })]);
    store.setSnapshot([elevator({ name: 'A' })]);
    expect(useLiveStore.getState().order).toEqual(['A']);
    expect(useLiveStore.getState().byName.B).toBeUndefined();
  });
});

describe('liveStore.setConnection', () => {
  beforeEach(() => useLiveStore.getState().reset());

  it('updates connection flags', () => {
    useLiveStore.getState().setConnection(true, 0);
    expect(useLiveStore.getState().connected).toBe(true);
  });
});
