// Package health provides a comprehensive health checking system for the elevator control system.
//
// This package implements multi-level health monitoring with the following key features:
//
// Health Check Types:
//   - Liveness checks: Verify the application is alive and responding
//   - Readiness checks: Ensure the application is ready to serve traffic
//   - System resource checks: Monitor memory usage and goroutine counts
//   - Component checks: Custom health checks for system components
//
// Key Components:
//   - HealthService: Central service that manages and coordinates all health checks
//   - HealthChecker: Interface for implementing custom health checks
//   - CheckResult: Standardized result format with status, timing, and details
//   - Caching: Built-in result caching with configurable TTL to prevent check overhead
//
// Status Levels:
//   - Healthy: Component is functioning normally
//   - Degraded: Component has issues but is still functional
//   - Unhealthy: Component is not functioning properly
//   - Unknown: Component status cannot be determined
//
// The health system supports parallel execution of checks, automatic caching,
// and provides detailed metrics and diagnostics for comprehensive system monitoring.
// This enables proactive issue detection, capacity planning, and system observability
// for the elevator control system in production environments.
//
// Usage:
//
//	service := NewHealthService(30 * time.Second)
//	service.Register(NewLivenessChecker())
//	service.Register(NewSystemResourceChecker(85.0, 1000))
//	status, results := service.GetOverallStatus(ctx)
package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// CheckResult represents the result of a health check
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Error     error                  `json:"error,omitempty"`
}

// HealthChecker defines the interface for health checks
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// HealthService manages all health checks
type HealthService struct {
	mu       sync.RWMutex
	checkers map[string]HealthChecker
	cache    map[string]CheckResult
	cacheTTL time.Duration
}

// NewHealthService creates a new health service
func NewHealthService(cacheTTL time.Duration) *HealthService {
	if cacheTTL <= 0 {
		cacheTTL = 30 * time.Second // Default cache TTL
	}

	return &HealthService{
		checkers: make(map[string]HealthChecker),
		cache:    make(map[string]CheckResult),
		cacheTTL: cacheTTL,
	}
}

// Register adds a health checker to the service
func (hs *HealthService) Register(checker HealthChecker) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.checkers[checker.Name()] = checker
}

// CheckAll runs all registered health checks
func (hs *HealthService) CheckAll(ctx context.Context) map[string]CheckResult {
	hs.mu.RLock()
	checkers := make(map[string]HealthChecker)
	for name, checker := range hs.checkers {
		checkers[name] = checker
	}
	hs.mu.RUnlock()

	results := make(map[string]CheckResult)
	resultCh := make(chan CheckResult, len(checkers))

	// Run checks in parallel
	var wg sync.WaitGroup
	for _, checker := range checkers {
		wg.Add(1)
		go func(c HealthChecker) {
			defer wg.Done()
			result := hs.checkWithCache(ctx, c)
			resultCh <- result
		}(checker)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for result := range resultCh {
		results[result.Name] = result
	}

	return results
}

// Check runs a specific health check
func (hs *HealthService) Check(ctx context.Context, name string) (CheckResult, error) {
	hs.mu.RLock()
	checker, exists := hs.checkers[name]
	hs.mu.RUnlock()

	if !exists {
		return CheckResult{}, fmt.Errorf("health checker '%s' not found", name)
	}

	return hs.checkWithCache(ctx, checker), nil
}

// checkWithCache checks if we have a cached result that's still valid
func (hs *HealthService) checkWithCache(ctx context.Context, checker HealthChecker) CheckResult {
	name := checker.Name()

	hs.mu.RLock()
	if cached, exists := hs.cache[name]; exists {
		if time.Since(cached.Timestamp) < hs.cacheTTL {
			hs.mu.RUnlock()
			return cached
		}
	}
	hs.mu.RUnlock()

	// Run the actual check
	result := checker.Check(ctx)

	// Cache the result
	hs.mu.Lock()
	hs.cache[name] = result
	hs.mu.Unlock()

	return result
}

// GetOverallStatus determines the overall system health
func (hs *HealthService) GetOverallStatus(ctx context.Context) (Status, map[string]CheckResult) {
	results := hs.CheckAll(ctx)

	overallStatus := StatusHealthy
	hasUnhealthy := false
	hasDegraded := false

	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		case StatusUnknown:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		overallStatus = StatusUnhealthy
	} else if hasDegraded {
		overallStatus = StatusDegraded
	}

	return overallStatus, results
}

// System Health Checkers

// SystemResourceChecker checks system resource utilization
type SystemResourceChecker struct {
	MemoryThresholdPercent float64
	GoroutineThreshold     int
}

func NewSystemResourceChecker(memThreshold float64, goroutineThreshold int) *SystemResourceChecker {
	if memThreshold <= 0 {
		memThreshold = 85.0 // Default 85% memory threshold
	}
	if goroutineThreshold <= 0 {
		goroutineThreshold = 1000 // Default goroutine threshold
	}

	return &SystemResourceChecker{
		MemoryThresholdPercent: memThreshold,
		GoroutineThreshold:     goroutineThreshold,
	}
}

func (src *SystemResourceChecker) Name() string {
	return "system_resources"
}

func (src *SystemResourceChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	numGoroutines := runtime.NumGoroutine()

	// Calculate memory usage percentage (simplified)
	memUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100

	status := StatusHealthy
	message := "System resources are healthy"
	details := map[string]interface{}{
		"memory_alloc_bytes":   m.Alloc,
		"memory_sys_bytes":     m.Sys,
		"memory_usage_percent": memUsagePercent,
		"goroutines":           numGoroutines,
		"gc_cycles":            m.NumGC,
		"heap_objects":         m.HeapObjects,
	}

	if memUsagePercent > src.MemoryThresholdPercent {
		status = StatusDegraded
		message = fmt.Sprintf("High memory usage: %.2f%%", memUsagePercent)
	}

	if numGoroutines > src.GoroutineThreshold {
		status = StatusDegraded
		if status == StatusDegraded {
			message += fmt.Sprintf(" and high goroutine count: %d", numGoroutines)
		} else {
			message = fmt.Sprintf("High goroutine count: %d", numGoroutines)
		}
	}

	return CheckResult{
		Name:      src.Name(),
		Status:    status,
		Message:   message,
		Details:   details,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}

// ComponentHealthChecker checks health of system components
type ComponentHealthChecker struct {
	componentName string
	healthFunc    func(ctx context.Context) (bool, string, map[string]interface{})
}

func NewComponentHealthChecker(name string, healthFunc func(ctx context.Context) (bool, string, map[string]interface{})) *ComponentHealthChecker {
	return &ComponentHealthChecker{
		componentName: name,
		healthFunc:    healthFunc,
	}
}

func (chc *ComponentHealthChecker) Name() string {
	return chc.componentName
}

func (chc *ComponentHealthChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	healthy, message, details := chc.healthFunc(ctx)

	status := StatusHealthy
	if !healthy {
		status = StatusUnhealthy
	}

	return CheckResult{
		Name:      chc.Name(),
		Status:    status,
		Message:   message,
		Details:   details,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}

// ReadinessChecker checks if the application is ready to serve traffic
type ReadinessChecker struct {
	dependencies []HealthChecker
}

func NewReadinessChecker(dependencies ...HealthChecker) *ReadinessChecker {
	return &ReadinessChecker{
		dependencies: dependencies,
	}
}

func (rc *ReadinessChecker) Name() string {
	return "readiness"
}

func (rc *ReadinessChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	status := StatusHealthy
	message := "Application is ready"
	details := make(map[string]interface{})

	unhealthyCount := 0
	for _, dep := range rc.dependencies {
		result := dep.Check(ctx)
		details[dep.Name()] = map[string]interface{}{
			"status":  result.Status,
			"message": result.Message,
		}

		if result.Status == StatusUnhealthy {
			unhealthyCount++
		}
	}

	if unhealthyCount > 0 {
		status = StatusUnhealthy
		message = fmt.Sprintf("Application not ready: %d unhealthy dependencies", unhealthyCount)
	}

	return CheckResult{
		Name:      rc.Name(),
		Status:    status,
		Message:   message,
		Details:   details,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}

// LivenessChecker checks if the application is alive and responding
type LivenessChecker struct {
	startTime time.Time
}

func NewLivenessChecker() *LivenessChecker {
	return &LivenessChecker{
		startTime: time.Now(),
	}
}

func (lc *LivenessChecker) Name() string {
	return "liveness"
}

func (lc *LivenessChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	uptime := time.Since(lc.startTime)

	return CheckResult{
		Name:    lc.Name(),
		Status:  StatusHealthy,
		Message: "Application is alive",
		Details: map[string]interface{}{
			"uptime_seconds": uptime.Seconds(),
			"start_time":     lc.startTime,
		},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}
}
