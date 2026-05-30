import { cn } from '@/shared/lib/cn';
import { useToastStore, type ToastKind } from '@/shared/store/toastStore';

const TONE: Record<ToastKind, string> = {
  info: 'bg-slate-800 text-white',
  success: 'bg-idle text-white',
  error: 'bg-fault text-white',
  warning: 'bg-deleting text-slate-900',
};

export function ToastContainer() {
  const toasts = useToastStore((s) => s.toasts);
  const dismiss = useToastStore((s) => s.dismiss);

  return (
    <div className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-80 flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          role="status"
          className={cn(
            'pointer-events-auto flex items-start justify-between gap-3 rounded-lg px-4 py-3 text-sm shadow-lg',
            TONE[t.kind],
          )}
        >
          <span>{t.message}</span>
          <button
            onClick={() => dismiss(t.id)}
            aria-label="Dismiss"
            className="opacity-70 transition-opacity hover:opacity-100"
          >
            ✕
          </button>
        </div>
      ))}
    </div>
  );
}
