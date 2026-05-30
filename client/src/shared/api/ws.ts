// WebSocket client for the live status stream.
//
// Differences from the old Svelte service (north-star design §4.4):
//   - NO connect() at import time. Lifecycle is owned by the caller (the
//     useBuildingLiveState hook), so connections are tied to a viewed building.
//   - Structured, level-gated logging (off by default) instead of console noise.
//   - Same exponential-backoff reconnection, kept because it works.

import { log } from '@/shared/lib/log';
import type { WireStatusBroadcast } from './wire';

type StatusListener = (raw: WireStatusBroadcast) => void;
type StateListener = (connected: boolean, retryCount: number) => void;

export interface ElevatorSocketOptions {
  url?: string;
  maxRetries?: number;
  baseDelayMs?: number;
}

export class ElevatorSocket {
  private ws: WebSocket | null = null;
  private retries = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private closedByClient = false;

  private readonly url: string;
  private readonly maxRetries: number;
  private readonly baseDelayMs: number;

  private statusListeners = new Set<StatusListener>();
  private stateListeners = new Set<StateListener>();

  constructor(opts: ElevatorSocketOptions = {}) {
    this.url = opts.url ?? import.meta.env.VITE_WS_URL ?? 'ws://localhost:6661/ws/status';
    this.maxRetries = opts.maxRetries ?? 6;
    this.baseDelayMs = opts.baseDelayMs ?? 1000;
  }

  onStatus(fn: StatusListener): () => void {
    this.statusListeners.add(fn);
    return () => this.statusListeners.delete(fn);
  }

  onStateChange(fn: StateListener): () => void {
    this.stateListeners.add(fn);
    return () => this.stateListeners.delete(fn);
  }

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return;
    }
    this.closedByClient = false;
    log.debug('ws: connecting', this.url);

    try {
      this.ws = new WebSocket(this.url);
    } catch (err) {
      log.error('ws: failed to construct socket', err);
      this.scheduleReconnect();
      return;
    }

    this.ws.onopen = () => {
      log.debug('ws: open');
      this.retries = 0;
      this.emitState(true);
    };

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data && typeof data === 'object' && !Array.isArray(data)) {
          this.statusListeners.forEach((fn) => fn(data as WireStatusBroadcast));
        }
      } catch (err) {
        log.error('ws: bad message', err);
      }
    };

    this.ws.onclose = () => {
      this.emitState(false);
      if (!this.closedByClient) this.scheduleReconnect();
    };

    this.ws.onerror = (err) => {
      log.error('ws: error', err);
    };
  }

  disconnect(): void {
    this.closedByClient = true;
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.ws) {
      this.ws.onclose = null; // avoid triggering reconnect on intentional close
      this.ws.close(1000, 'client disconnect');
      this.ws = null;
    }
    this.emitState(false);
  }

  private scheduleReconnect(): void {
    if (this.retries >= this.maxRetries) {
      log.error('ws: max reconnect attempts reached');
      return;
    }
    const delay = this.baseDelayMs * 2 ** this.retries;
    this.retries += 1;
    this.emitState(false);
    log.debug(`ws: reconnect in ${delay}ms (attempt ${this.retries})`);
    this.reconnectTimer = setTimeout(() => this.connect(), delay);
  }

  private emitState(connected: boolean): void {
    this.stateListeners.forEach((fn) => fn(connected, this.retries));
  }
}
