import { cn } from '@/shared/lib/cn';
import { Button } from '@/shared/ui/Button';
import { useThemeStore } from '@/shared/store/themeStore';

interface HeaderProps {
  connected: boolean;
  retryCount: number;
  elevatorCount: number;
  onCreateClick: () => void;
}

export function Header({ connected, retryCount, elevatorCount, onCreateClick }: HeaderProps) {
  const { mode, toggle } = useThemeStore();

  return (
    <header className="flex items-center justify-between border-b border-slate-200 bg-white px-5 py-3 dark:border-slate-700 dark:bg-slate-800">
      <div className="flex items-center gap-3">
        <span className="text-lg font-bold text-slate-900 dark:text-white">Elevator Control</span>
        <span className="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/40 dark:text-blue-300">
          React
        </span>
      </div>

      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-xs" aria-live="polite">
          <span
            className={cn(
              'h-2.5 w-2.5 rounded-full',
              connected ? 'bg-idle' : retryCount > 0 ? 'bg-deleting animate-pulse' : 'bg-fault',
            )}
            aria-hidden
          />
          <span className="text-slate-600 dark:text-slate-300">
            {connected ? 'Live' : retryCount > 0 ? `Reconnecting (${retryCount})` : 'Disconnected'}
          </span>
        </div>

        <span className="text-xs text-slate-500 dark:text-slate-400">
          {elevatorCount} elevator{elevatorCount === 1 ? '' : 's'}
        </span>

        <Button variant="ghost" size="sm" onClick={toggle} aria-label="Toggle theme">
          {mode === 'dark' ? '☀️' : '🌙'}
        </Button>

        <Button size="sm" onClick={onCreateClick}>
          + Elevator
        </Button>
      </div>
    </header>
  );
}
