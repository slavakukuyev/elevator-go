import { useState } from 'react';
import { Header } from './Header';
import { BuildingView } from '@/features/building/components/BuildingView';
import { CreateElevatorModal } from '@/features/controls/CreateElevatorModal';
import { ToastContainer } from '@/shared/ui/ToastContainer';
import { useBuildingLiveState } from '@/features/building/useBuildingLiveState';
import { CURRENT_BUILDING_ID } from '@/features/building/constants';

/**
 * Project 1 shell: a single building. The building id is a constant here; in
 * Project 2 it comes from a route param and the shell gains a building selector.
 */
export function App() {
  const [createOpen, setCreateOpen] = useState(false);

  // Header + create-modal need connection/fleet info. The BuildingView consumes
  // the same store via its own selector subscription, so this is not a second
  // socket — useBuildingLiveState's effect dedupes through the shared store.
  const { elevators, connected, retryCount } = useBuildingLiveState(CURRENT_BUILDING_ID);
  const existingNames = elevators.map((e) => e.name);

  return (
    <div className="flex h-full flex-col bg-slate-50 dark:bg-slate-900">
      <Header
        connected={connected}
        retryCount={retryCount}
        elevatorCount={elevators.length}
        onCreateClick={() => setCreateOpen(true)}
      />

      <main className="min-h-0 flex-1">
        <BuildingView buildingId={CURRENT_BUILDING_ID} onCreateClick={() => setCreateOpen(true)} />
      </main>

      <CreateElevatorModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        existingNames={existingNames}
      />
      <ToastContainer />
    </div>
  );
}
