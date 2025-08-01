package manager_benchmarks

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
	"github.com/slavakukuyev/elevator-go/internal/manager"
)

// buildManagerTestConfig creates a test configuration for benchmarks
func buildManagerTestConfig() *config.Config {
	return &config.Config{
		LogLevel:              "ERROR", // Reduce logging noise in benchmarks
		Port:                  8080,
		MinFloor:              -10,
		MaxFloor:              50,
		EachFloorDuration:     time.Millisecond * 10,
		OpenDoorDuration:      time.Millisecond * 10,
		RequestTimeout:        time.Second * 30, // Increased for concurrent benchmarks
		CreateElevatorTimeout: time.Second * 20, // Increased for elevator creation
		OperationTimeout:      time.Second * 60, // Increased for long operations
		StatusUpdateTimeout:   time.Second * 10, // Increased for status updates
		HealthCheckTimeout:    time.Second * 5,  // Increased for health checks
	}
}

// BenchmarkManager_AddElevator benchmarks elevator addition performance
func BenchmarkManager_AddElevator(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		elevatorName := fmt.Sprintf("BenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 50, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_RequestElevator benchmarks elevator request processing performance
func BenchmarkManager_RequestElevator(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	for i := 0; i < 5; i++ {
		elevatorName := fmt.Sprintf("BenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 100, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		from := i % 90
		to := from + 10
		if to > 100 {
			to = 100
		}

		_, err := mgr.RequestElevator(ctx, from, to)
		if err != nil {
			b.Logf("Request failed: %v", err)
		}
	}
}

// BenchmarkManager_ConcurrentRequests benchmarks concurrent request handling
func BenchmarkManager_ConcurrentRequests(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	for i := 0; i < 10; i++ {
		elevatorName := fmt.Sprintf("ConcurrentBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 100, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			from := counter % 90
			to := from + 10
			if to > 100 {
				to = 100
			}

			_, err := mgr.RequestElevator(ctx, from, to)
			if err != nil {
				// Log but don't fail - some requests may legitimately fail
				b.Logf("Request failed: %v", err)
			}
			counter++
		}
	})
}

// BenchmarkManager_GetElevators benchmarks elevator retrieval performance
func BenchmarkManager_GetElevators(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	for i := 0; i < 50; i++ {
		elevatorName := fmt.Sprintf("GetBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 20, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = mgr.GetElevators()
	}
}

// BenchmarkManager_GetElevator benchmarks single elevator retrieval by name
func BenchmarkManager_GetElevator(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	elevatorNames := make([]string, 50)
	for i := 0; i < 50; i++ {
		elevatorName := fmt.Sprintf("GetBenchmarkElevator%d", i)
		elevatorNames[i] = elevatorName
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 20, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		name := elevatorNames[i%len(elevatorNames)]
		_ = mgr.GetElevator(name)
	}
}

// BenchmarkManager_ElevatorSelection benchmarks the elevator selection algorithm through the public API
func BenchmarkManager_ElevatorSelection(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup multiple elevators with varying ranges
	for i := 0; i < 20; i++ {
		elevatorName := fmt.Sprintf("SelectionBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 100, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	// Benchmark the actual elevator selection through RequestElevator
	for i := 0; i < b.N; i++ {
		from := 50 + (i % 5) // Vary the starting floor slightly
		to := from + 10
		if to > 100 {
			to = 100
		}

		_, err := mgr.RequestElevator(ctx, from, to)
		if err != nil {
			b.Logf("Request failed: %v", err)
		}
	}
}

// BenchmarkManager_GetStatus benchmarks status retrieval performance
func BenchmarkManager_GetStatus(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators with some requests
	for i := 0; i < 10; i++ {
		elevatorName := fmt.Sprintf("StatusBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 50, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}

		// Add some requests to make status more complex
		_, _ = mgr.RequestElevator(ctx, i, i+10)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mgr.GetStatus()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_GetHealthStatus benchmarks health status retrieval
func BenchmarkManager_GetHealthStatus(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	for i := 0; i < 5; i++ {
		elevatorName := fmt.Sprintf("HealthBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 50, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := mgr.GetHealthStatus()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkManager_GetMetrics benchmarks metrics retrieval
func BenchmarkManager_GetMetrics(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup elevators
	for i := 0; i < 5; i++ {
		elevatorName := fmt.Sprintf("MetricsBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 50, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = mgr.GetMetrics()
	}
}

// BenchmarkManager_MemoryUsage benchmarks memory usage under load
func BenchmarkManager_MemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		cfg := buildManagerTestConfig()
		elevatorFactory := &factory.StandardElevatorFactory{}
		mgr := manager.New(cfg, elevatorFactory)

		// Create multiple elevators and add requests
		for j := 0; j < 5; j++ {
			elevatorName := fmt.Sprintf("MemoryBenchmarkElevator%d_%d", i, j)
			err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 50, time.Millisecond*10, time.Millisecond*10, 12)
			if err != nil {
				b.Fatal(err)
			}

			// Add requests
			for k := 0; k < 5; k++ {
				_, _ = mgr.RequestElevator(ctx, k, k+10)
			}
		}

		// Access various properties
		_ = mgr.GetElevators()
		_, _ = mgr.GetStatus()
	}
}

// BenchmarkManager_ConcurrentMixed benchmarks mixed concurrent operations
func BenchmarkManager_ConcurrentMixed(b *testing.B) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	elevatorFactory := &factory.StandardElevatorFactory{}
	mgr := manager.New(cfg, elevatorFactory)

	// Setup initial elevators
	for i := 0; i < 5; i++ {
		elevatorName := fmt.Sprintf("MixedBenchmarkElevator%d", i)
		err := mgr.AddElevator(ctx, cfg, elevatorName, 0, 100, time.Millisecond*10, time.Millisecond*10, 12)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			switch counter % 4 {
			case 0:
				// Request elevator
				from := counter % 90
				to := from + 10
				if to > 100 {
					to = 100
				}
				_, _ = mgr.RequestElevator(ctx, from, to)
			case 1:
				// Get elevators
				_ = mgr.GetElevators()
			case 2:
				// Get status
				_, _ = mgr.GetStatus()
			case 3:
				// Get specific elevator
				elevatorName := fmt.Sprintf("MixedBenchmarkElevator%d", counter%5)
				_ = mgr.GetElevator(elevatorName)
			}
			counter++
		}
	})
}
