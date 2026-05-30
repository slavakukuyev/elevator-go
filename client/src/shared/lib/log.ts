// Level-gated logger. Off by default — set VITE_LOG=debug to enable.
// Replaces the scattered console.log noise in the old client (north-star design §4.4).

type Level = 'debug' | 'info' | 'warn' | 'error' | 'silent';

const ORDER: Record<Level, number> = { debug: 0, info: 1, warn: 2, error: 3, silent: 4 };

const configured = (import.meta.env.VITE_LOG as Level | undefined) ?? 'warn';
const threshold = ORDER[configured] ?? ORDER.warn;

function emit(level: Exclude<Level, 'silent'>, args: unknown[]) {
  if (ORDER[level] < threshold) return;
  console[level === 'debug' ? 'log' : level](...args);
}

export const log = {
  debug: (...args: unknown[]) => emit('debug', args),
  info: (...args: unknown[]) => emit('info', args),
  warn: (...args: unknown[]) => emit('warn', args),
  error: (...args: unknown[]) => emit('error', args),
};
