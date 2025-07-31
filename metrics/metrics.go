package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/slavakukuyev/elevator-go/internal/constants"
)

var (
	// Request processing metrics
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricsNamespace + "_request_duration_seconds",
			Help:    "Duration of elevator request processing",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{constants.ElevatorNameLabel, "status"},
	)

	// Request counters
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricsNamespace + "_requests_total",
			Help: "Total number of elevator requests",
		},
		[]string{constants.ElevatorNameLabel, "direction", "status"},
	)

	// Elevator efficiency metrics
	elevatorEfficiency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_efficiency_ratio",
			Help: "Elevator efficiency ratio (successful requests / total requests)",
		},
		[]string{constants.ElevatorNameLabel},
	)

	// Wait time tracking
	waitTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricsNamespace + "_wait_time_seconds",
			Help:    "Time passengers wait for elevator arrival",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{constants.ElevatorNameLabel},
	)

	// Elevator travel metrics
	travelTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricsNamespace + "_travel_time_seconds",
			Help:    "Time for elevator to complete a journey",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
		},
		[]string{constants.ElevatorNameLabel, "floors_traveled"},
	)

	// System health metrics
	systemHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_system_health",
			Help: "System health status (1 = healthy, 0 = unhealthy)",
		},
		[]string{"component"},
	)

	// Current system state
	currentFloor = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_current_floor",
			Help: "Current floor of each elevator",
		},
		[]string{constants.ElevatorNameLabel},
	)

	pendingRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_pending_requests",
			Help: "Number of pending requests per elevator",
		},
		[]string{constants.ElevatorNameLabel, "direction"},
	)

	// Circuit breaker metrics
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_circuit_breaker_state",
			Help: "Circuit breaker state (0 = closed, 1 = half-open, 2 = open)",
		},
		[]string{constants.ElevatorNameLabel},
	)

	circuitBreakerFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricsNamespace + "_circuit_breaker_failures_total",
			Help: "Total circuit breaker failures",
		},
		[]string{constants.ElevatorNameLabel},
	)

	// HTTP metrics
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    constants.MetricsNamespace + "_http_request_duration_seconds",
			Help:    "HTTP request duration",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricsNamespace + "_http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	// Error rate metrics
	errorRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: constants.MetricsNamespace + "_errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"error_type", "component"},
	)

	// Performance metrics
	avgResponseTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_avg_response_time_seconds",
			Help: "Average response time for system operations",
		},
		[]string{"operation"},
	)

	// Resource utilization
	memoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"type"},
	)

	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: constants.MetricsNamespace + "_active_connections",
			Help: "Number of active WebSocket connections",
		},
	)
)

func init() {
	prometheus.MustRegister(
		requestDuration,
		requestsTotal,
		elevatorEfficiency,
		waitTime,
		travelTime,
		systemHealth,
		currentFloor,
		pendingRequests,
		circuitBreakerState,
		circuitBreakerFailures,
		httpRequestDuration,
		httpRequestsTotal,
		errorRate,
		avgResponseTime,
		memoryUsage,
		activeConnections,
	)
}

// Request processing metrics
func RecordRequestDuration(elevatorName, status string, seconds float64) {
	requestDuration.WithLabelValues(elevatorName, status).Observe(seconds)
}

func IncRequestsTotal(elevatorName, direction, status string) {
	requestsTotal.WithLabelValues(elevatorName, direction, status).Inc()
}

// Elevator efficiency metrics
func SetElevatorEfficiency(elevatorName string, ratio float64) {
	elevatorEfficiency.WithLabelValues(elevatorName).Set(ratio)
}

func RecordWaitTime(elevatorName string, seconds float64) {
	waitTime.WithLabelValues(elevatorName).Observe(seconds)
}

func RecordTravelTime(elevatorName, floorsTraveled string, seconds float64) {
	travelTime.WithLabelValues(elevatorName, floorsTraveled).Observe(seconds)
}

// System health metrics
func SetSystemHealth(component string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	systemHealth.WithLabelValues(component).Set(value)
}

// Current state metrics
func SetCurrentFloor(elevatorName string, floor float64) {
	currentFloor.WithLabelValues(elevatorName).Set(floor)
}

func SetPendingRequests(elevatorName, direction string, count float64) {
	pendingRequests.WithLabelValues(elevatorName, direction).Set(count)
}

// Circuit breaker metrics
func SetCircuitBreakerState(elevatorName string, state float64) {
	circuitBreakerState.WithLabelValues(elevatorName).Set(state)
}

func IncCircuitBreakerFailures(elevatorName string) {
	circuitBreakerFailures.WithLabelValues(elevatorName).Inc()
}

// HTTP metrics
func RecordHTTPRequest(method, endpoint, statusCode string, seconds float64) {
	httpRequestDuration.WithLabelValues(method, endpoint, statusCode).Observe(seconds)
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

// Error tracking
func IncError(errorType, component string) {
	errorRate.WithLabelValues(errorType, component).Inc()
}

// Performance metrics
func SetAvgResponseTime(operation string, seconds float64) {
	avgResponseTime.WithLabelValues(operation).Set(seconds)
}

// Resource utilization
func SetMemoryUsage(memType string, bytes float64) {
	memoryUsage.WithLabelValues(memType).Set(bytes)
}

func SetActiveConnections(count float64) {
	activeConnections.Set(count)
}

// Legacy function for backward compatibility
func RequestDurationHistogram(elevatorName string, seconds float64) {
	RecordRequestDuration(elevatorName, "success", seconds)
}
