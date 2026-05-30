import { useMemo } from 'react';
import { cn } from '@/shared/lib/cn';
import { computeFleetStats, statusShare } from '../domain/stats';
import type { Elevator, ElevatorStatus } from '../domain/types';

interface FleetSummaryProps {
  elevators: Elevator[];
}

const ROWS: { status: ElevatorStatus; label: string; dot: string; bar: string }[] = [
  { status: 'idle', label: 'Idle', dot: 'bg-idle', bar: 'bg-idle' },
  { status: 'moving', label: 'Moving', dot: 'bg-moving', bar: 'bg-moving' },
  { status: 'deleting', label: 'Removing', dot: 'bg-deleting', bar: 'bg-deleting' },
  { status: 'fault', label: 'Fault', dot: 'bg-fault', bar: 'bg-fault' },
];

/**
 * Compact, always-visible fleet summary. Replaces the Svelte MonitoringDashboard
 * (counts + utilization bars + pending requests) but driven entirely by the live
 * WebSocket snapshot — no extra polling or REST sync.
 */
export function FleetSummary({ elevators }: FleetSummaryProps) {
  const stats = useMemo(() => computeFleetStats(elevators), [elevators]);

  return (
    <aside
      aria-label="Fleet summary"
      className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-700 dark:bg-slate-800"
    >
      <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500 dark:text-slate-400">
        Fleet summary
      </h2>

      <div className="mb-4 flex items-baseline gap-2">
        <span className="text-3xl font-bold text-slate-900 dark:text-white">{stats.total}</span>
        <span className="text-sm text-slate-500 dark:text-slate-400">
          elevator{stats.total === 1 ? '' : 's'}
        </span>
        <span className="ml-auto text-sm text-slate-500 dark:text-slate-400">
          {stats.pendingRequests} pending request{stats.pendingRequests === 1 ? '' : 's'}
        </span>
      </div>

      <ul className="space-y-2.5">
        {ROWS.filter((r) => r.status !== 'fault' || stats.fault > 0).map((row) => {
          const share = statusShare(stats, row.status);
          return (
            <li key={row.status}>
              <div className="mb-1 flex items-center justify-between text-xs">
                <span className="flex items-center gap-2 text-slate-600 dark:text-slate-300">
                  <span className={cn('h-2.5 w-2.5 rounded-full', row.dot)} aria-hidden />
                  {row.label}
                </span>
                <span className="font-medium text-slate-900 dark:text-white">
                  {stats[row.status]}
                </span>
              </div>
              <div className="h-1.5 w-full overflow-hidden rounded-full bg-slate-100 dark:bg-slate-700">
                <div
                  className={cn('h-full rounded-full transition-[width] duration-500', row.bar)}
                  style={{ width: `${share}%` }}
                  aria-hidden
                />
              </div>
            </li>
          );
        })}
      </ul>
    </aside>
  );
}
