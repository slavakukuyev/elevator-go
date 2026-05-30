// Backend wire shapes — the exact JSON the Go backend sends/accepts.
// Kept isolated so the rest of the app never touches snake_case or wrapper envelopes.
//
// Source of truth: internal/http/handlers.go, internal/domain/elevator_status.go.
// (north-star design §6: candidate for OpenAPI-generated types in P1→P2.)

/** All /v1 REST responses are wrapped in this envelope. */
export interface V1Response<T> {
  success: boolean;
  data: T;
  meta: { request_id: string; version: string; duration: string };
  timestamp: string;
}

/** Structured error envelope returned on non-2xx. */
export interface V1ErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    details?: string;
    request_id: string;
    user_message?: string;
  };
  meta: { request_id: string; version: string; duration: string };
  timestamp: string;
}

export interface WireElevatorCreate {
  name: string;
  min_floor: number;
  max_floor: number;
  message: string;
}

export interface WireFloorRequest {
  elevator_name: string;
  from_floor: number;
  to_floor: number;
  direction: string;
  message: string;
}

export interface WireHealth {
  status: string;
  timestamp: string;
  checks: {
    total_elevators?: number;
    system_healthy?: boolean;
  };
}

/**
 * WebSocket status broadcast: a map keyed by elevator name.
 * direction is one of "up" | "down" | "" (idle) | "deleting".
 */
export type WireStatusBroadcast = Record<string, WireElevatorStatus>;

export interface WireElevatorStatus {
  name: string;
  current_floor: number;
  direction: string;
  requests: number;
  min_floor: number;
  max_floor: number;
  is_deleting: boolean;
}
