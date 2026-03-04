# CLAUDE.md

## Key Commands

```bash
make server-dev          # Backend local (ports 6660/6661, auto-rebuild)
make client-dev          # Frontend Vite dev (port 5173)
make dev/local           # Both locally (recommended)
make build               # Production binary
make test/unit           # Unit tests (fast)
make test/race           # Race detection
make test/acceptance     # E2E tests
make test/all            # Full suite
make lint                # Go + TypeScript linters
make lint/go             # golangci-lint
make lint/ts             # ESLint on .ts/.svelte
make lint/fix            # Auto-fix TS + Prettier
make docker/compose      # Full stack via nginx (port 8080)
make cleanup             # Kill ports 6660, 6661, 5173
```

## Architecture

Clean Architecture: **Handlers → Manager → Elevator → Domain** (unidirectional)

- `internal/elevator/` — SCAN/LOOK algorithm, goroutine-based movement → [`docs/elevator.md`](docs/elevator.md)
- `internal/manager/` — fleet coordination, 3-phase selection, load balancing → [`docs/manager.md`](docs/manager.md)
- `internal/http/` — REST `/v1/`, WebSocket `/ws/status`
- `internal/domain/` — types and interfaces
- `internal/infra/` — config, observability, logging
- `client/` — Svelte 4 + TypeScript + Tailwind

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/elevators` | Create elevator |
| DELETE | `/v1/elevators` | Graceful delete → [`docs/elevator_deletion.md`](docs/elevator_deletion.md) |
| POST | `/v1/floors/request` | Request floor service |
| GET | `/v1/health` | Health check |
| GET | `/v1/metrics` | Prometheus metrics |
| WS | `/ws/status` | Real-time status broadcast |

## Key Patterns

**Concurrency**: `sync.Mutex` for shared state, `atomic.Bool` for flags (e.g. `isDeleting`), always propagate `context.Context`

**Error handling**: `fmt.Errorf("context: %w", err)`, structured `slog` with correlation IDs

**Testing**: table-driven, parallel, mocks in `tests/mocks/`, `make test/race` to verify

**Observability** → [`docs/logging_and_error_handling.md`](docs/logging_and_error_handling.md), [`docs/metrics.md`](docs/metrics.md)

**Concurrency perf** → [`docs/CONCURRENCY_PERFORMANCE.md`](docs/CONCURRENCY_PERFORMANCE.md)

## Identifiers

Elevators are identified by `name` (no UUID). Name is the primary key throughout manager, API, and WebSocket. See [`docs/identifier_design.md`](docs/identifier_design.md).

## Elevator Deletion

Graceful: elevator finishes all queued requests before removal. `isDeleting atomic.Bool` flag prevents new requests without stopping SCAN movement. See [`docs/elevator_deletion.md`](docs/elevator_deletion.md).

## Environment

```bash
ENV=development LOG_LEVEL=DEBUG DEFAULT_ELEVATOR_COUNT=0  # dev / acceptance tests
ENV=production  LOG_LEVEL=INFO  DEFAULT_ELEVATOR_COUNT=3  # prod
```

Config reference → [`docs/configuration.md`](docs/configuration.md)

## Load Thresholds (hardcoded, TODO: make configurable)

Normal ≤8 requests, Moderate 9–12, Overload >12

## Testing Gotchas

- Acceptance tests require `DEFAULT_ELEVATOR_COUNT=0`
- Integration tests require Docker (testcontainers)
- Race tests are slow — use targeted, not in CI hot path
