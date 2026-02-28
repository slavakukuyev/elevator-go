SERVER=cmd/server/main.go
BIN_PATH=out/bin
BIN_NAME=elevator
PKGS=./...
UNIT_PKGS=$(shell go list ./internal/... ./cmd/...)

# Docker configuration
DOCKER_IMAGE_NAME=elevator-service
DOCKER_TAG=latest
BUILD_DIR=build

# Port configuration
BACKEND_PORTS=6660,6661
CLIENT_PORT=5173
ALL_PORTS=$(BACKEND_PORTS),$(CLIENT_PORT)

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m

# Port cleanup function
define cleanup_ports
	@echo "Cleaning up ports: $(1)..."
	@if lsof -ti:$(1) >/dev/null 2>&1; then \
		lsof -ti:$(1) | xargs kill -9 2>/dev/null || true; \
		echo "Cleaned up processes on ports: $(1)"; \
	else \
		echo "No processes running on ports: $(1)"; \
	fi
endef

.PHONY: help
help:
	@echo "$(GREEN)Elevator Control System - Available Commands$(NC)"
	@echo ""
	@echo "$(YELLOW)Build & Run:$(NC)"
	@echo "  build               - Build the elevator server"
	@echo "  clean               - Clean build artifacts"
	@echo "  run                 - Build and run the elevator server"
	@echo ""
	@echo "$(YELLOW)Development:$(NC)"
	@echo "  dev/client          - Run client in dev mode only"
	@echo "  dev/local           - Run both backend and client locally"
	@echo "  server-dev          - Run backend server locally"
	@echo ""
	@echo "$(YELLOW)Docker:$(NC)"
	@echo "  docker/build        - Build Docker image"
	@echo "  docker/run          - Run container (backend only)"
	@echo "  docker/compose      - Run full compose setup"
	@echo "  docker/stop         - Stop and clean Docker containers"
	@echo "  dev/full            - Run backend (Docker) + client (dev mode)"
	@echo "  dev/backend         - Run backend in Docker only"
	@echo "  dev/stop            - Stop all development services"
	@echo ""
	@echo "$(YELLOW)Testing:$(NC)"
	@echo "  test/unit           - Run unit tests"
	@echo "  test/race           - Run tests with race detection"
	@echo "  test/acceptance     - Run acceptance tests"
	@echo "  test/integration    - Run integration tests"
	@echo "  test/benchmarks     - Run benchmark tests"
	@echo "  test/all            - Run all tests"
	@echo ""
	@echo "$(YELLOW)Utilities:$(NC)"
	@echo "  cleanup             - Clean up all ports and processes"
	@echo "  debug-prepare       - Prepare environment for debugging"

# Build targets
.PHONY: build build_server clean run
build: build_server

build_server:
	go build -o ${BIN_PATH}/${BIN_NAME} ${SERVER}

clean:
	rm -rf ${BIN_PATH}

run: build_server
	./${BIN_PATH}/${BIN_NAME}

# Development targets
.PHONY: cleanup server-dev client-dev dev/full dev/backend dev/client dev/local dev/stop

# Unified cleanup target
cleanup:
	$(call cleanup_ports,$(ALL_PORTS))

server-dev: 
	$(call cleanup_ports,$(BACKEND_PORTS))
	@echo "Building and starting Go server locally..."
	@make build_server
	@echo "Backend will be available at: http://localhost:6660"
	@echo "WebSocket will be available at: http://localhost:6661" 
	ENV=development LOG_LEVEL=DEBUG DEFAULT_ELEVATOR_COUNT=0 ./${BIN_PATH}/${BIN_NAME}

client-dev:
	$(call cleanup_ports,$(CLIENT_PORT))
	@echo "Starting client dev server on port 5173..."
	cd client && npm run dev

dev/full:
	@echo "Starting backend in Docker and client in dev mode..."
	docker-compose -f docker-compose.full.yml up backend -d
	@echo "Waiting for backend to be ready..."
	@sleep 10
	@make client-dev

dev/backend:
	@echo "Starting backend in Docker..."
	@echo "Backend will be available at: http://localhost:6660"
	docker-compose -f docker-compose.full.yml up backend

dev/client: client-dev

dev/local:
	@echo "Starting local development environment..."
	@make server-dev &
	@echo "Waiting for backend to be ready..."
	@sleep 5
	@make client-dev

dev/stop:
	@echo "Stopping all development services..."
	docker-compose -f docker-compose.full.yml down 2>/dev/null || true
	$(call cleanup_ports,$(ALL_PORTS))

# Docker targets
.PHONY: docker/build docker/run docker/compose docker/stop

docker/build:
	@echo "Building Docker image..."
	docker build -f ${BUILD_DIR}/package/Dockerfile -t ${DOCKER_IMAGE_NAME}:${DOCKER_TAG} .

docker/run: docker/build
	@echo "Running Docker container..."
	docker run --rm -p 6660:6660 \
		-e ENV=development \
		-e LOG_LEVEL=DEBUG \
		-e DEFAULT_ELEVATOR_COUNT=3 \
		--name ${DOCKER_IMAGE_NAME} \
		${DOCKER_IMAGE_NAME}:${DOCKER_TAG}

docker/compose:
	@echo "Starting Docker Compose setup..."
	@echo "Web interface: http://localhost:8080"
	@echo "API: http://localhost:6660"
	@echo "WebSocket: http://localhost:6661"
	docker-compose up -d

docker/stop:
	@echo "Stopping all Docker services..."
	@docker stop ${DOCKER_IMAGE_NAME} 2>/dev/null || true
	@docker rm ${DOCKER_IMAGE_NAME} 2>/dev/null || true
	@docker-compose down 2>/dev/null || true
	@docker-compose -f docker-compose.full.yml down 2>/dev/null || true
	@docker system prune -f

# Test targets
.PHONY: test/unit test/race test/acceptance test/integration test/benchmarks test/all

test/unit:
	go test -v -short $(UNIT_PKGS)

test/race:
	@echo "Running race tests with packages: $(UNIT_PKGS)"
	@echo "Race tests started at: $(shell date)"
	GORACE=1 go test -v -race $(UNIT_PKGS)
	@echo "Race tests completed at: $(shell date)"

test/race-full:
	@echo "Running full race tests (without -short flag) with packages: $(UNIT_PKGS)"
	@echo "Full race tests started at: $(shell date)"
	go test -v -race $(UNIT_PKGS)
	@echo "Full race tests completed at: $(shell date)"

test/acceptance:
	go test -v ./tests/acceptance/acceptance_test.go -timeout 60s

test/integration:
	go test -v ./tests/acceptance/acceptance_testcontainers_test.go -timeout 360s

test/benchmarks:
	go test -v -bench=. ./tests/benchmarks/...

test/all: test/unit test/race test/acceptance test/integration
	@echo "$(GREEN)All tests completed successfully!$(NC)"

# Debug preparation (simplified)
.PHONY: debug-prepare
debug-prepare: cleanup
	@echo "$(GREEN)Environment prepared for IDE debugging.$(NC)"
	@echo "Available ports:"
	@echo "  - HTTP API: 6660"
	@echo "  - WebSocket: 6661"
	@echo "  - Client: 5173"
