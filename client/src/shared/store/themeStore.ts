import { create } from 'zustand';

type Mode = 'light' | 'dark';

function initialMode(): Mode {
  if (typeof window === 'undefined') return 'light';
  const stored = window.localStorage.getItem('theme');
  if (stored === 'light' || stored === 'dark') return stored;
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

function apply(mode: Mode) {
  if (typeof document === 'undefined') return;
  document.documentElement.classList.toggle('dark', mode === 'dark');
  window.localStorage.setItem('theme', mode);
}

interface ThemeState {
  mode: Mode;
  toggle: () => void;
}

export const useThemeStore = create<ThemeState>((set, get) => {
  const mode = initialMode();
  apply(mode);
  return {
    mode,
    toggle: () => {
      const next: Mode = get().mode === 'dark' ? 'light' : 'dark';
      apply(next);
      set({ mode: next });
    },
  };
});
