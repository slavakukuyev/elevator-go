# Elevator Control System

A Go-based elevator control system with real-time WebSocket updates and Svelte frontend.

## Architecture

### Backend (Go)
- **Elevator Engine**: SCAN/LOOK algorithm implementation with direction-based request management
- **Manager System**: Multi-elevator coordination with load balancing and capacity management
- **HTTP API**: RESTful endpoints with versioning (`/v1/`) and comprehensive error handling
- **WebSocket Server**: Real-time status broadcasting on port 6661
- **Circuit Breaker**: Fault tolerance implementation for elevator operations

### Frontend (Svelte)
- **Real-time Visualization**: Live elevator position and state updates
- **Control Panel**: Elevator creation and floor request management
- **Monitoring Dashboard**: System metrics and health status
- **Responsive Design**: Mobile-friendly interface with dark mode support

## Current Implementation Status

### ✅ Implemented Features

#### Core Elevator Logic
- SCAN/LOOK algorithm with intelligent direction switching
- Multi-elevator fleet management (up to 100 elevators)
- Load balancing and overload protection
- Support for negative floors (underground parking)
- Configurable floor ranges and timing parameters

#### API Endpoints
- `POST /v1/elevators` - Create new elevator
- `POST /v1/floors/request` - Request elevator service
- `GET /v1/health` - System health status
- `GET /v1/metrics` - Performance metrics
- `GET /v1` - API information
- Legacy endpoints: `/elevator`, `/floor`, `/health`, `/metrics/system`

#### WebSocket
- Real-time status updates on `/ws/status`
- JSON-formatted elevator state broadcasts
- Connection management with ping/pong

#### Observability
- Structured JSON logging with correlation IDs
- Prometheus metrics collection
- OpenTelemetry tracing support
- Health check endpoints (`/v1/health/live`, `/v1/health/ready`)

#### Configuration
- Environment-based configuration (development, testing, production)
- Comprehensive timeout and rate limiting settings
- Circuit breaker configuration
- WebSocket connection management

### 🔧 Technical Stack

#### Backend
- **Go 1.24+** with Go modules
- **HTTP/WebSocket** servers on ports 6660/6661
- **Prometheus** metrics collection
- **OpenTelemetry** for distributed tracing
- **Structured logging** with slog

#### Frontend
- **Svelte 4** with TypeScript
- **Tailwind CSS** for styling
- **Vite** for development and building
- **WebSocket** client for real-time updates

#### Development Tools
- **Docker** containerization
- **Make** build automation
- **Comprehensive testing** (unit, integration, acceptance)
- **ESLint/Prettier** code formatting

## Quick Start

### Prerequisites
- Go 1.24+
- Node.js 18+ (for frontend)
- Docker (optional)

### Local Development

```bash
# Clone repository
git clone <repository-url>
cd elevator

# Backend only
make server-dev

# Frontend only  
make client-dev

# Full stack (backend in Docker + frontend dev)
make dev/full
```

### Docker Deployment

```bash
# Full stack with nginx
make docker/compose

# Backend only
make docker/run
```

### API Usage

```bash
# Create elevator
curl -X POST http://localhost:6660/v1/elevators \
  -H "Content-Type: application/json" \
  -d '{"name": "Elevator-1", "min_floor": -2, "max_floor": 20}'

# Request elevator
curl -X POST http://localhost:6660/v1/floors/request \
  -H "Content-Type: application/json" \
  -d '{"from_floor": 1, "to_floor": 10}'

# Health check
curl http://localhost:6660/v1/health
```

## Configuration

### Environment Variables

```bash
# Core settings
ENV=development
PORT=6660
LOG_LEVEL=INFO

# Elevator configuration
DEFAULT_MAX_FLOOR=9
DEFAULT_MIN_FLOOR=0
DEFAULT_OVERLOAD_THRESHOLD=12
EACH_FLOOR_DURATION=500ms
OPEN_DOOR_DURATION=2s

# Performance tuning
RATE_LIMIT_RPM=100
CIRCUIT_BREAKER_MAX_FAILURES=5
WEBSOCKET_MAX_CONNECTIONS=1000
```

### Environment-Specific Configs
- **Development**: Enhanced debugging, detailed logging
- **Testing**: Fast operations, minimal resources
- **Production**: Optimized performance, security hardening

## Testing

### Test Coverage
- **Unit Tests**: Component-level testing with 90%+ coverage
- **Integration Tests**: End-to-end API validation
- **Acceptance Tests**: Real-world scenario simulation
- **Benchmark Tests**: Performance testing

### Test Commands
```bash
make test/unit          # Unit tests
make test/acceptance    # Acceptance tests
make test/benchmarks    # Performance benchmarks
make test/all           # All tests
```

## Project Structure

```
elevator/
├── cmd/server/          # Application entry point
├── internal/            # Core application logic
│   ├── elevator/        # Elevator algorithm implementation
│   ├── manager/         # Fleet management and coordination
│   ├── http/           # HTTP server and API handlers
│   ├── domain/         # Business logic and types
│   └── infra/          # Infrastructure and configuration
├── client/             # Svelte frontend application
├── configs/            # Environment-specific configurations
├── tests/              # Test suites and utilities
└── docs/               # Documentation
```

## Development Workflow

```bash
# Run tests
make test/unit

# Run benchmarks
make test/benchmarks

# Build for production
make build

# Run with specific environment
ENV=testing make run
```

## Monitoring & Observability

### Metrics Endpoints
- `/metrics` - Prometheus-compatible metrics
- `/v1/health` - System health status
- `/ws/status` - Real-time WebSocket updates

### Key Metrics
- Request rate and response times
- Elevator utilization and efficiency
- Error rates and system health
- Resource usage (memory, CPU)

### Logging
- Structured JSON format
- Correlation IDs for request tracing
- Environment-specific log levels

## API Documentation

### V1 Endpoints

#### Create Elevator
```http
POST /v1/elevators
Content-Type: application/json

{
  "name": "Elevator-1",
  "min_floor": -2,
  "max_floor": 20
}
```

#### Request Elevator
```http
POST /v1/floors/request
Content-Type: application/json

{
  "from_floor": 1,
  "to_floor": 10
}
```

#### Health Check
```http
GET /v1/health
```

#### Metrics
```http
GET /v1/metrics
```

## Performance Characteristics

### Current Benchmarks
- **Throughput**: ~1000 requests/minute per elevator
- **Response Time**: <5ms request processing
- **Elevator Selection**: <300ms average
- **Memory Usage**: ~50MB base + ~2MB per elevator
- **CPU Usage**: <5% idle, <20% under load

### Scalability
- **Max Elevators**: 100 per system
- **Concurrent Requests**: 1000+ per minute
- **WebSocket Connections**: 1000 max
- **Memory Scaling**: Linear with elevator count

## Known Limitations

### Current Constraints
- Single-instance deployment (no clustering)
- In-memory state (no persistence)
- Limited elevator capacity modeling
- No advanced scheduling algorithms

### Planned Improvements
- Database persistence layer
- Multi-instance clustering
- Advanced traffic prediction
- Energy efficiency optimization

## Contributing

### Development Guidelines
- Follow Go best practices and clean architecture
- Maintain comprehensive test coverage
- Use structured logging and proper error handling
- Document public APIs and complex algorithms

### Code Quality
- **Linting**: `golangci-lint` for Go code
- **Formatting**: `go fmt` and `goimports`
- **Testing**: Comprehensive test suites
- **Documentation**: GoDoc comments

## License

MIT License - see LICENSE file for details.
