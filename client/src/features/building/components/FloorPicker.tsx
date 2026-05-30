import { useMemo } from 'react';
import { Modal } from '@/shared/ui/Modal';
import { Button } from '@/shared/ui/Button';
import { cn } from '@/shared/lib/cn';
import { formatFloor, reachableDestinations } from '../domain/floors';
import type { Elevator } from '../domain/types';

interface FloorPickerProps {
  fromFloor: number | null;
  elevators: Elevator[];
  onSelect: (to: number) => void;
  onClose: () => void;
  pending: boolean;
}

function floorTone(floor: number): string {
  if (floor === 0) return 'border-deleting/50 bg-deleting/10';
  if (floor < 0) return 'border-moving/40 bg-moving/10';
  return 'border-slate-300 bg-slate-50 dark:border-slate-600 dark:bg-slate-700/50';
}

export function FloorPicker({ fromFloor, elevators, onSelect, onClose, pending }: FloorPickerProps) {
  const destinations = useMemo(
    () => (fromFloor === null ? [] : reachableDestinations(fromFloor, elevators).sort((a, b) => b - a)),
    [fromFloor, elevators],
  );

  return (
    <Modal
      open={fromFloor !== null}
      title={fromFloor === null ? 'Select destination' : `Trip from floor ${formatFloor(fromFloor)}`}
      onClose={onClose}
      size="sm"
    >
      {destinations.length === 0 ? (
        <p className="text-sm text-slate-500 dark:text-slate-400">No reachable destinations.</p>
      ) : (
        <div className="grid grid-cols-5 gap-2">
          {destinations.map((floor) => (
            <button
              key={floor}
              disabled={pending}
              onClick={() => onSelect(floor)}
              aria-label={`Go to floor ${formatFloor(floor)}`}
              className={cn(
                'aspect-square rounded-md border text-sm font-bold text-slate-900 transition-transform hover:scale-105 disabled:opacity-50 dark:text-white',
                floorTone(floor),
              )}
            >
              {formatFloor(floor)}
            </button>
          ))}
        </div>
      )}
      <div className="mt-4 flex justify-end">
        <Button variant="ghost" size="sm" onClick={onClose}>
          Cancel
        </Button>
      </div>
    </Modal>
  );
}
