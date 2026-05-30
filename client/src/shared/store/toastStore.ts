import { create } from 'zustand';

export type ToastKind = 'info' | 'success' | 'error' | 'warning';

export interface Toast {
  id: string;
  message: string;
  kind: ToastKind;
}

interface ToastState {
  toasts: Toast[];
  push: (message: string, kind?: ToastKind, durationMs?: number) => void;
  dismiss: (id: string) => void;
}

let seq = 0;

export const useToastStore = create<ToastState>((set, get) => ({
  toasts: [],
  push: (message, kind = 'info', durationMs = 5000) => {
    const id = `t${seq++}`;
    set((s) => ({ toasts: [...s.toasts, { id, message, kind }] }));
    if (durationMs > 0) {
      setTimeout(() => get().dismiss(id), durationMs);
    }
  },
  dismiss: (id) => set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}));

/** Imperative helper for non-component callers (e.g. mutation onError). */
export const toast = {
  success: (m: string) => useToastStore.getState().push(m, 'success'),
  error: (m: string) => useToastStore.getState().push(m, 'error', 7000),
  info: (m: string) => useToastStore.getState().push(m, 'info'),
  warning: (m: string) => useToastStore.getState().push(m, 'warning'),
};
