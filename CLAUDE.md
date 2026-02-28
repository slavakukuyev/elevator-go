# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Backend Development (Go)
```bash
make server-dev          # Run backend locally (auto-rebuilds, port 6660/6661)
make build               # Build production binary
make run                 # Build and run server
make cleanup             # Kill processes on ports 6660, 6661, 5173
```

### Frontend Development (Svelte)
```bash
make client-dev          # Run Vite dev server (port 5173)
cd client && npm run build     # Production build
cd client && npm run lint      # ESLint check
cd client && npm run format    # Prettier formatting
```

### Full Stack Development
```bash
make dev/local           # Backend + frontend locally (recommended)
make dev/full            # Backend in Docker + frontend local dev
make dev/backend         # Backend in Docker only
make dev/stop            # Stop all dev services
```

### Testing
```bash
make test/unit           # Unit tests only (fast)
make test/race           # Race condition detection
make test/acceptance     # End-to-end acceptance tests
make test/integration    # Integration tests with testcontainers
make test/benchmarks     # Performance benchmarks
make test/all            # Full test suite
```

### Docker
```bash
make docker/build        # Build Docker image
make docker/run          # Run backend in container
make docker/compose      # Full stack with nginx (port 8080)
make docker/stop         # Stop and clean all containers
```

## Architecture Overview

### High-Level Design
This is a **multi-elevator control system** with real-time WebSocket updates, following **Clean Architecture** principles:

```
┌─────────────────────────────────────────────────┐
│ HTTP API (port 6660) + WebSocket (port 6661)   │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│            internal/http/                       │
│  - handlers.go: API endpoints                   │
│  - middleware.go: Request validation, CORS      │
│  - websocket.go: Real-time status broadcasts    │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│          internal/manager/                      │
│  - Manager: Multi-elevator fleet coordinator    │
│  - Elevator selection algorithm (3-phase):      │
│    1. Idle elevators (preferred)                │
│    2. Same direction elevators                  │
│    3. Opposite direction with load balancing    │
│  - Capacity management and overload protection  │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│         internal/elevator/                      │
│  - SCAN/LOOK algorithm implementation           │
│  - Direction-based request management           │
│  - Floor transitions and door operations        │
│  - Concurrent request handling with mutex       │
└──────────────────┬──────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────┐
│          internal/domain/                       │
│  - Elevator, Request, ElevatorState types       │
│  - Business logic interfaces                    │
└─────────────────────────────────────────────────┘
```

### Core Components

#### 1. **Elevator Algorithm** (`internal/elevator/`)
- **SCAN/LOOK algorithm**: Intelligent direction-based request servicing
- **Key Feature**: Handles underground parking (negative floors) and high-rises
- **Concurrency**: Goroutine-based movement with mutex-protected state
- **Critical Files**:
  - `elevator.go`: Main Run() loop and state management
  - `requests.go`: Request queuing and filtering
  - See `docs/elevator.md` for detailed algorithm explanation

#### 2. **Manager System** (`internal/manager/`)
- **3-Phase Selection Algorithm**:
  1. Prefers idle elevators (lowest distance + load score)
  2. Same-direction elevators with capacity check
  3. Opposite-direction with load balancing
- **Load Classification**:
  - Normal: ≤8 requests
  - Moderate: 9-12 requests
  - Overload: >12 requests
- **Circuit Breaker**: Fault tolerance for elevator operations
- **Critical Files**: `manager.go`, see `docs/manager.md`

#### 3. **HTTP API & WebSocket** (`internal/http/`)
- **RESTful API** (versioned at `/v1/`)
  - `POST /v1/elevators` - Create elevator
  - `POST /v1/floors/request` - Request service
  - `GET /v1/health` - Health checks
  - `GET /v1/metrics` - Prometheus metrics
- **WebSocket**: `/ws/status` - Real-time elevator state broadcasts
- **Middleware**: Request validation, CORS, timeout handling

#### 4. **Observability** (`internal/infra/observability/`)
- **OpenTelemetry**: Distributed tracing with context propagation
- **Prometheus**: Request metrics, latency, elevator utilization
- **Structured Logging**: JSON logs with correlation IDs (slog)
- **Health Checks**: `/v1/health/live` (liveness), `/v1/health/ready` (readiness)

#### 5. **Configuration** (`internal/infra/config/`)
- **Environment-based**: development, testing, production configs in `configs/`
- **Key Settings**:
  - Floor ranges: `DEFAULT_MIN_FLOOR`, `DEFAULT_MAX_FLOOR`
  - Timings: `EACH_FLOOR_DURATION` (500ms), `OPEN_DOOR_DURATION` (2s)
  - Limits: `RATE_LIMIT_RPM`, `CIRCUIT_BREAKER_MAX_FAILURES`
  - WebSocket: `WEBSOCKET_MAX_CONNECTIONS` (1000)

### Frontend Architecture (`client/`)
- **Svelte 4** + **TypeScript** + **Tailwind CSS**
- **SvelteKit** for routing and server-side rendering
- **Real-time Updates**: WebSocket client in `src/services/websocket.js`
- **State Management**: Svelte stores in `src/stores/`
- **Key Components**:
  - `src/components/ElevatorPanel.svelte` - Control interface
  - `src/components/ElevatorVisualization.svelte` - Real-time display
  - `src/routes/+page.svelte` - Main dashboard

## Code Standards & Patterns

### Backend (Go)

**Clean Architecture Rules**:
- **Handlers → Manager → Elevator → Domain** (unidirectional dependency)
- Always use **interfaces** for public functions, not concrete types
- **Dependency injection** via constructor functions
- **Context propagation** for request tracing and cancellation

**Concurrency Safety**:
- Guard shared state with `sync.Mutex` or channels
- Always propagate `context.Context` for goroutine cancellation
- Use `defer` for resource cleanup
- Test with `make test/race` to detect race conditions

**Error Handling**:
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Log errors with structured fields (correlation IDs)
- Return meaningful HTTP status codes

**Testing**:
- Table-driven tests with parallel execution
- Mock external interfaces (see `tests/mocks/`)
- Separate unit tests (fast) from integration tests (slow)
- Aim for 90%+ coverage on `internal/` packages

**Observability**:
- Start OpenTelemetry spans for all service boundaries
- Include `context.Context` in all logs and traces
- Record important span attributes (request params, errors)
- Use correlation IDs for request tracing

### Frontend (Svelte)

**Code Review Before Changes**:
- Always review existing code in `<CODE_REVIEW>` tags
- Create implementation plan in `<PLANNING>` tags
- Preserve variable names and string literals unless changing intentionally

**Security**:
- Validate all user inputs (especially floor numbers)
- Use `<SECURITY_REVIEW>` tags for auth or input handling changes
- Sanitize data before rendering

**Incremental Development**:
- Break changes into small, testable steps
- Suggest testing after each stage
- Ask for clarification on ambiguous requirements

### DevOps

**Pre-Implementation Review**:
- Review infrastructure code in `<CODE_REVIEW>` tags
- Document changes in `<PLANNING>` tags
- Use `<SECURITY_REVIEW>` for security-sensitive changes

**Production Changes**:
- Always ask for clarification before production modifications
- Consider operational impact (monitoring, maintenance)
- Test incrementally with small changes

## Important Implementation Notes

### Known TODO Items
- Manager overload thresholds are hardcoded (8, 12, 15 requests) - should be configurable
- Wait time estimation uses fixed constants (2s/floor, 3s/door)
- Circuit breaker configuration could be more granular

### Testing Gotchas
- Acceptance tests expect zero elevators at startup: set `DEFAULT_ELEVATOR_COUNT=0`
- Integration tests use testcontainers: requires Docker running
- Race tests are slow: use `test/race` for specific checks, not in CI

### Environment Setup
```bash
# Development
ENV=development LOG_LEVEL=DEBUG DEFAULT_ELEVATOR_COUNT=0

# Testing
ENV=testing LOG_LEVEL=INFO DEFAULT_ELEVATOR_COUNT=0

# Production
ENV=production LOG_LEVEL=INFO DEFAULT_ELEVATOR_COUNT=3
```

### Debugging
```bash
make debug-prepare       # Clean ports for IDE debugging
# Then run main.go from IDE with ENV=development
```

## Tech Stack Summary

### Backend
- **Go 1.25+** (module: `github.com/slavakukuyev/elevator-go`)
- **gorilla/websocket** - WebSocket server
- **prometheus/client_golang** - Metrics
- **OpenTelemetry** - Distributed tracing
- **testcontainers-go** - Integration testing

### Frontend
- **Svelte 4** with **SvelteKit 1.20+**
- **TypeScript 5.0** - Type safety
- **Vite 4** - Build tool and dev server
- **Tailwind CSS 3** - Styling
- **Vitest** - Unit testing

### Infrastructure
- **Docker** + **docker-compose**
- **nginx** for reverse proxy
- **Prometheus** for metrics collection (optional)

## Additional Documentation

- `docs/elevator.md` - Detailed SCAN/LOOK algorithm explanation
- `docs/manager.md` - Elevator selection and load balancing
- `docs/configuration.md` - Environment configuration guide
- `docs/logging_and_error_handling.md` - Observability patterns
- `docs/CONCURRENCY_PERFORMANCE.md` - Concurrency design and benchmarks
- `README.md` - User-facing documentation and quick start
