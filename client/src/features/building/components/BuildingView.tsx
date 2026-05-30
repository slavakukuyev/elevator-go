import { useMemo, useState } from 'react';
import { useBuildingLiveState } from '../useBuildingLiveState';
import { useRequestFloor, useDeleteElevator, useLoadSampleFleet } from '../mutations';
import { ElevatorShaft } from './ElevatorShaft';
import { FloorPicker } from './FloorPicker';
import { FleetSummary } from './FleetSummary';
import { StatusLegend } from './StatusLegend';
import { Button } from '@/shared/ui/Button';
import { ConfirmDialog } from '@/shared/ui/ConfirmDialog';
import type { Elevator } from '../domain/types';

interface BuildingViewProps {
  buildingId: string;
  onCreateClick: () => void;
}

/** Split fleet into main vs parking, mirroring the original grouping. */
function partition(elevators: Elevator[]) {
  const main: Elevator[] = [];
  const parking: Elevator[] = [];
  for (const e of elevators) {
    (e.name.toLowerCase().includes('parking') ? parking : main).push(e);
  }
  return { main, parking };
}

export function BuildingView({ buildingId, onCreateClick }: BuildingViewProps) {
  const { elevators, connected } = useBuildingLiveState(buildingId);
  const requestFloor = useRequestFloor();
  const deleteElevator = useDeleteElevator();
  const loadSample = useLoadSampleFleet();

  const [fromFloor, setFromFloor] = useState<number | null>(null);
  const [pendingDelete, setPendingDelete] = useState<string | null>(null);

  const { main, parking } = useMemo(() => partition(elevators), [elevators]);
  const existingNames = useMemo(() => elevators.map((e) => e.name), [elevators]);

  const handleSelectDestination = (to: number) => {
    if (fromFloor === null) return;
    requestFloor.mutate({ from: fromFloor, to }, { onSuccess: () => setFromFloor(null) });
  };

  const confirmDelete = () => {
    if (pendingDelete === null) return;
    deleteElevator.mutate(pendingDelete);
    setPendingDelete(null);
  };

  const renderShaft = (e: Elevator) => (
    <ElevatorShaft key={e.name} elevator={e} onPickFloor={setFromFloor} onDelete={setPendingDelete} />
  );

  // Distinguish "still connecting" from "connected but empty" so the empty-state
  // CTA doesn't flash before the first WebSocket snapshot arrives.
  if (elevators.length === 0 && !connected) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="flex flex-col items-center gap-3 text-slate-500 dark:text-slate-400">
          <span
            className="h-8 w-8 animate-spin rounded-full border-2 border-slate-300 border-t-blue-500 dark:border-slate-600 dark:border-t-blue-400"
            aria-hidden
          />
          <p className="text-sm">Connecting to the building…</p>
        </div>
      </div>
    );
  }

  if (elevators.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-6">
        <div className="max-w-sm text-center">
          <div className="mx-auto mb-5 flex h-20 w-20 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800">
            <svg className="h-10 w-10 text-slate-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
              />
            </svg>
          </div>
          <h2 className="mb-2 text-xl font-bold text-slate-900 dark:text-white">Building ready</h2>
          <p className="mb-5 text-sm text-slate-500 dark:text-slate-400">
            Create your first elevator, or load a sample fleet to explore the simulation.
          </p>
          <div className="flex justify-center gap-3">
            <Button onClick={onCreateClick}>+ Create elevator</Button>
            <Button
              variant="secondary"
              onClick={() => loadSample.mutate(existingNames)}
              loading={loadSample.isPending}
            >
              Load sample fleet
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-auto p-6">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-bold text-slate-900 dark:text-white">Building overview</h1>
        <StatusLegend />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_18rem]">
        <div>
          {main.length > 0 && (
            <section className="mb-8">
              <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
                Main elevators
              </h2>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 2xl:grid-cols-3">
                {main.map(renderShaft)}
              </div>
            </section>
          )}

          {parking.length > 0 && (
            <section>
              <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
                Parking elevators
              </h2>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 2xl:grid-cols-3">
                {parking.map(renderShaft)}
              </div>
            </section>
          )}
        </div>

        <FleetSummary elevators={elevators} />
      </div>

      <FloorPicker
        fromFloor={fromFloor}
        elevators={elevators}
        onSelect={handleSelectDestination}
        onClose={() => setFromFloor(null)}
        pending={requestFloor.isPending}
      />

      <ConfirmDialog
        open={pendingDelete !== null}
        title="Delete elevator"
        message={
          pendingDelete
            ? `Delete elevator "${pendingDelete}"? It will finish its queued trips, then be removed.`
            : ''
        }
        confirmLabel="Delete"
        destructive
        onConfirm={confirmDelete}
        onCancel={() => setPendingDelete(null)}
      />
    </div>
  );
}
