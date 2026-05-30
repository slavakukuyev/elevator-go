// Floor helpers — pure functions, ported and trimmed from the Svelte client's
// FloorSelectionService. Only what the UI actually uses is kept.

import type { Elevator } from './types';

/** Format a floor for display: 0→"G", negatives→"B{n}", else the number. */
export function formatFloor(floor: number | undefined): string {
  if (floor === undefined || floor === null) return '?';
  if (floor === 0) return 'G';
  if (floor < 0) return `B${Math.abs(floor)}`;
  return String(floor);
}

/** Floors served by at least one operational elevator, descending (top→bottom). */
export function buildingFloorsDescending(elevators: Elevator[]): number[] {
  if (elevators.length === 0) return [];
  const min = Math.min(...elevators.map((e) => e.minFloor));
  const max = Math.max(...elevators.map((e) => e.maxFloor));
  return Array.from({ length: max - min + 1 }, (_, i) => max - i);
}

function isOperational(e: Elevator): boolean {
  return e.status !== 'fault' && e.status !== 'deleting';
}

/**
 * Destination floors reachable from `fromFloor`: any floor within range of an
 * operational elevator that also serves `fromFloor`. Sorted ascending.
 */
export function reachableDestinations(fromFloor: number, elevators: Elevator[]): number[] {
  const dests = new Set<number>();
  for (const e of elevators) {
    if (!isOperational(e)) continue;
    if (fromFloor < e.minFloor || fromFloor > e.maxFloor) continue;
    for (let f = e.minFloor; f <= e.maxFloor; f++) {
      if (f !== fromFloor) dests.add(f);
    }
  }
  return [...dests].sort((a, b) => a - b);
}

/** Elevators that physically serve a given floor (range check only). */
export function elevatorsServingFloor(floor: number, elevators: Elevator[]): Elevator[] {
  return elevators.filter((e) => floor >= e.minFloor && floor <= e.maxFloor);
}
