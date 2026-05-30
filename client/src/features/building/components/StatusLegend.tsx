import { cn } from '@/shared/lib/cn';

const ITEMS: { label: string; dot: string }[] = [
  { label: 'Idle', dot: 'bg-idle' },
  { label: 'Moving', dot: 'bg-moving' },
  { label: 'Removing', dot: 'bg-deleting' },
];

/** Inline key explaining what the elevator-car colors mean. */
export function StatusLegend() {
  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-slate-500 dark:text-slate-400">
      {ITEMS.map((item) => (
        <span key={item.label} className="flex items-center gap-1.5">
          <span className={cn('h-2.5 w-2.5 rounded-full', item.dot)} aria-hidden />
          {item.label}
        </span>
      ))}
    </div>
  );
}
