// Typed REST client. Exposes ONLY endpoints the backend actually implements
// (north-star design §4.4 — the dead "Not implemented" stubs are gone).

import type { ElevatorConfig, FloorRequest, SystemHealth } from '@/features/building/domain/types';
import type {
  V1Response,
  V1ErrorResponse,
  WireElevatorCreate,
  WireFloorRequest,
  WireHealth,
} from './wire';

const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:6660/v1';

/** Normalized API error carrying the backend's structured error envelope when present. */
export class ApiError extends Error {
  code: string;
  userMessage?: string;
  details?: string;
  requestId?: string;
  status: number;

  constructor(init: {
    message: string;
    code?: string;
    userMessage?: string;
    details?: string;
    requestId?: string;
    status: number;
  }) {
    super(init.message);
    this.name = 'ApiError';
    this.code = init.code ?? 'UNKNOWN_ERROR';
    this.userMessage = init.userMessage;
    this.details = init.details;
    this.requestId = init.requestId;
    this.status = init.status;
  }

  /** The best human-facing string to show the user. */
  get display(): string {
    return this.userMessage || this.message;
  }
}

async function request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${endpoint}`, {
    headers: { 'Content-Type': 'application/json', ...options.headers },
    ...options,
  });

  const text = await res.text();
  const body = text ? JSON.parse(text) : {};

  if (!res.ok) {
    const err = body as Partial<V1ErrorResponse>;
    throw new ApiError({
      message: err.error?.message ?? `API error ${res.status}`,
      code: err.error?.code,
      userMessage: err.error?.user_message,
      details: err.error?.details,
      requestId: err.error?.request_id,
      status: res.status,
    });
  }

  return body as T;
}

export const elevatorApi = {
  async createElevator(config: ElevatorConfig): Promise<void> {
    await request<V1Response<WireElevatorCreate>>('/elevators', {
      method: 'POST',
      body: JSON.stringify({
        name: config.name,
        min_floor: config.minFloor,
        max_floor: config.maxFloor,
      }),
    });
  },

  async deleteElevator(name: string): Promise<void> {
    await request<V1Response<unknown>>('/elevators', {
      method: 'DELETE',
      body: JSON.stringify({ name }),
    });
  },

  async requestFloor({ from, to }: FloorRequest): Promise<void> {
    await request<V1Response<WireFloorRequest>>('/floors/request', {
      method: 'POST',
      body: JSON.stringify({ from, to }),
    });
  },

  async getHealth(): Promise<SystemHealth> {
    const res = await request<V1Response<WireHealth>>('/health');
    return {
      healthy: res.data.status === 'healthy',
      elevatorCount: res.data.checks.total_elevators ?? 0,
    };
  },
};
