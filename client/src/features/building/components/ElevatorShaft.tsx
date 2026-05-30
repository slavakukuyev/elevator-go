import { memo, useMemo } from 'react';
import { cn } from '@/shared/lib/cn';
import { formatFloor } from '../domain/floors';
import { FLOOR_PX } from '../constants';
import type { Elevator, ElevatorStatus } from '../domain/types';

interface ElevatorShaftProps {
  elevator: Elevator;
  onPickFloor: (floor: number) => void;
  onDelete: (name: string) => void;
}

const STATUS_CAR: Record<ElevatorStatus, string> = {
  idle: 'bg-idle',
  moving: 'bg-moving',
  deleting: 'bg-deleting',
  fault: 'bg-fault',
};

const STATUS_DOT: Record<ElevatorStatus, string> = {
  idle: 'bg-idle',
  moving: 'bg-moving',
  deleting: 'bg-deleting',
  fault: 'bg-fault',
};

const STATUS_LABEL: Record<ElevatorStatus, string> = {
  idle: 'Idle',
  moving: 'Moving',
  deleting: 'Removing',
  fault: 'Fault',
};

function directionArrow(direction: Elevator['direction']): string {
  if (direction === 'up') return '↑';
  if (direction === 'down') return '↓';
  return '';
}

/** Vertical offset (px) of a floor row from the top of the shaft. */
function floorTop(floor: number, minFloor: number, maxFloor: number): number {
  const totalFloors = maxFloor - minFloor + 1;
  const indexFromTop = totalFloors - (floor - minFloor) - 1;
  return indexFromTop * FLOOR_PX;
}

function ElevatorShaftBase({ elevator, onPickFloor, onDelete }: ElevatorShaftProps) {
  const { name, minFloor, maxFloor, currentFloor, status, direction } = elevator;

  const floors = useMemo(
    () => Array.from({ length: maxFloor - minFloor + 1 }, (_, i) => maxFloor - i),
    [minFloor, maxFloor],
  );

  const shaftHeight = floors.length * FLOOR_PX;
  const carTop = floorTop(currentFloor, minFloor, maxFloor);
  const showsZeroLine = minFloor <= 0 && maxFloor >= 0;

  return (
    <section
      className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition-shadow hover:shadow-md dark:border-slate-700 dark:bg-slate-800"
      aria-label={`Elevator ${name}`}
    >
      {/* Header */}
      <div className="mb-3 flex items-start justify-between">
        <div>
          <h3 className="font-semibold text-slate-900 dark:text-white">{name}</h3>
          <p className="text-xs text-slate-500 dark:text-slate-400">
            Floors {formatFloor(minFloor)}–{formatFloor(maxFloor)}
          </p>
        </div>
        {status === 'deleting' ? (
          <span className="text-xs font-medium text-deleting">Removing…</span>
        ) : (
          <button
            onClick={() => onDelete(name)}
            aria-label={`Delete elevator ${name}`}
            title="Delete elevator"
            className="rounded p-1 text-slate-400 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/30"
          >
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
              />
            </svg>
          </button>
        )}
      </div>

      {/* Status row */}
      <div className="mb-3 flex items-center gap-2 text-xs">
        <span className={cn('h-2.5 w-2.5 rounded-full', STATUS_DOT[status])} aria-hidden />
        <span className="font-medium text-slate-700 dark:text-slate-200">{STATUS_LABEL[status]}</span>
        {direction && <span className="font-bold text-slate-700 dark:text-slate-200">{directionArrow(direction)}</span>}
        {elevator.requests > 0 && (
          <span className="ml-auto text-slate-500 dark:text-slate-400">
            {elevator.requests} request{elevator.requests > 1 ? 's' : ''}
          </span>
        )}
      </div>

      {/* Shaft */}
      <div className="relative flex gap-2 rounded-lg bg-slate-100 p-2 dark:bg-slate-900/50">
        {/* Floor labels (clickable to request a trip from that floor) */}
        <div className="relative w-9" style={{ height: shaftHeight }}>
          {floors.map((floor) => (
            <button
              key={floor}
              onClick={() => onPickFloor(floor)}
              aria-label={`Request a trip from floor ${formatFloor(floor)}`}
              style={{ top: floorTop(floor, minFloor, maxFloor), height: FLOOR_PX }}
              className={cn(
                'absolute flex w-full items-center justify-center rounded text-xs font-medium transition-colors hover:bg-slate-200 dark:hover:bg-slate-700',
                floor === 0
                  ? 'text-deleting'
                  : floor === currentFloor
                    ? 'font-bold text-idle'
                    : 'text-slate-500 dark:text-slate-400',
              )}
            >
              {formatFloor(floor)}
            </button>
          ))}
        </div>

        {/* Car track */}
        <div
          className="relative flex-1 rounded border-2 border-slate-300 dark:border-slate-600"
          style={{ height: shaftHeight }}
        >
          {/* Floor gridlines */}
          {floors.map((floor) => (
            <div
              key={floor}
              className={cn(
                'absolute left-0 right-0 h-px',
                floor === 0 ? 'bg-deleting/70' : 'bg-slate-200 dark:bg-slate-700',
              )}
              style={{ top: floorTop(floor, minFloor, maxFloor) + FLOOR_PX / 2 }}
              aria-hidden
            />
          ))}
          {showsZeroLine && null}

          {/* The car — CSS transition animates between per-second backend snapshots */}
          <div
            data-testid="elevator-car"
            data-current-floor={currentFloor}
            data-status={status}
            className={cn(
              'absolute left-1 right-1 flex flex-col items-center justify-center rounded-md text-white shadow-car',
              'transition-[top] duration-700 ease-car',
              STATUS_CAR[status],
              status === 'moving' && 'ring-2 ring-white/40',
            )}
            style={{ top: carTop + 2, height: FLOOR_PX - 4 }}
          >
            <span className="text-sm font-bold leading-none">{formatFloor(currentFloor)}</span>
            {direction && <span className="text-base leading-none">{directionArrow(direction)}</span>}
          </div>
        </div>
      </div>
    </section>
  );
}

// Memoize: with preserved object identity from the store, a row only re-renders
// when its own elevator actually changed.
export const ElevatorShaft = memo(ElevatorShaftBase);
