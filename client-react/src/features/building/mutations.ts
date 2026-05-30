// Command hooks (create / delete / request floor) via TanStack Query.
// Live state arrives over WebSocket, so these mutations don't need to write the
// store — they just fire the command and surface success/error as toasts.

import { useMutation } from '@tanstack/react-query';
import { elevatorApi, ApiError } from '@/shared/api/client';
import { toast } from '@/shared/store/toastStore';
import { formatFloor } from './domain/floors';
import type { ElevatorConfig, FloorRequest } from './domain/types';

function describeError(err: unknown): string {
  if (err instanceof ApiError) return err.display;
  if (err instanceof Error) return err.message;
  return 'Unexpected error';
}

export function useCreateElevator() {
  return useMutation({
    mutationFn: (config: ElevatorConfig) => elevatorApi.createElevator(config),
    onSuccess: (_d, config) => toast.success(`Elevator "${config.name}" created`),
    onError: (err) => toast.error(describeError(err)),
  });
}

export function useDeleteElevator() {
  return useMutation({
    mutationFn: (name: string) => elevatorApi.deleteElevator(name),
    onSuccess: (_d, name) => toast.info(`Elevator "${name}" will be removed when idle`),
    onError: (err) => toast.error(describeError(err)),
  });
}

export function useRequestFloor() {
  return useMutation({
    mutationFn: (req: FloorRequest) => elevatorApi.requestFloor(req),
    onSuccess: (_d, req) =>
      toast.success(`Requested ${formatFloor(req.from)} → ${formatFloor(req.to)}`),
    onError: (err) => toast.error(describeError(err)),
  });
}

// A small, varied demo fleet — mirrors the Svelte client's "Load sample data"
// button (main + parking elevators with mixed floor ranges). Created via the
// real backend; the live WS stream then renders them.
const SAMPLE_FLEET: ElevatorConfig[] = [
  { name: 'Main-A', minFloor: -2, maxFloor: 10 },
  { name: 'Main-B', minFloor: -1, maxFloor: 8 },
  { name: 'Main-C', minFloor: 0, maxFloor: 12 },
  { name: 'Parking-I', minFloor: -5, maxFloor: 2 },
  { name: 'Parking-J', minFloor: -3, maxFloor: 1 },
];

export function useLoadSampleFleet() {
  return useMutation({
    mutationFn: async (existingNames: string[]) => {
      const fresh = SAMPLE_FLEET.filter((c) => !existingNames.includes(c.name));
      let created = 0;
      for (const config of fresh) {
        try {
          await elevatorApi.createElevator(config);
          created += 1;
        } catch {
          // Skip an already-existing or rejected config; keep going with the rest.
        }
      }
      return created;
    },
    onSuccess: (created) =>
      created > 0
        ? toast.success(`Created ${created} sample elevator${created === 1 ? '' : 's'}`)
        : toast.info('Sample fleet already present'),
    onError: (err) => toast.error(describeError(err)),
  });
}
