// Ephemeral live state for the currently-viewed building.
//
// Zustand chosen over Context+reducer because the WS pushes a full snapshot
// ~1×/s and we want fine-grained selector subscriptions so only changed
// elevators re-render (north-star design §4.4 / §6).

import { create } from 'zustand';
import type { Elevator } from '@/features/building/domain/types';

export interface LiveState {
  /** Elevators keyed by name for O(1) diffing. */
  byName: Record<string, Elevator>;
  order: string[];
  connected: boolean;
  retryCount: number;

  setSnapshot: (elevators: Elevator[]) => void;
  setConnection: (connected: boolean, retryCount: number) => void;
  reset: () => void;
}

/** Shallow-equality check so unchanged elevators keep their object identity. */
function sameElevator(a: Elevator | undefined, b: Elevator): boolean {
  return (
    !!a &&
    a.currentFloor === b.currentFloor &&
    a.status === b.status &&
    a.direction === b.direction &&
    a.requests === b.requests &&
    a.minFloor === b.minFloor &&
    a.maxFloor === b.maxFloor
  );
}

export const useLiveStore = create<LiveState>((set) => ({
  byName: {},
  order: [],
  connected: false,
  retryCount: 0,

  setSnapshot: (elevators) =>
    set((state) => {
      const next: Record<string, Elevator> = {};
      let changed = elevators.length !== state.order.length;

      for (const e of elevators) {
        const prev = state.byName[e.name];
        // Preserve identity when nothing changed → memoized rows skip re-render.
        next[e.name] = sameElevator(prev, e) ? prev : e;
        if (next[e.name] !== prev) changed = true;
      }

      const order = elevators.map((e) => e.name);
      if (!changed && order.join() === state.order.join()) return state;
      return { byName: next, order };
    }),

  setConnection: (connected, retryCount) =>
    set((state) =>
      state.connected === connected && state.retryCount === retryCount
        ? state
        : { connected, retryCount },
    ),

  reset: () => set({ byName: {}, order: [], connected: false, retryCount: 0 }),
}));
