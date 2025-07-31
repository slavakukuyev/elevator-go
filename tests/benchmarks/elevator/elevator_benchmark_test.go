package elevator_benchmarks

import (
	"testing"
	"time"

	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/elevator"
)

// BenchmarkElevator_New benchmarks elevator creation performance
func BenchmarkElevator_New(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		elev, err := elevator.New("BenchmarkElevator", 0, 50, time.Millisecond*100, time.Millisecond*50,
			30*time.Second, 5, 30*time.Second, 3, 12)
		if err != nil {
			b.Fatal(err)
		}
		// Properly cleanup the elevator to prevent goroutine leaks
		elev.Shutdown()
	}
}

// BenchmarkElevator_Request benchmarks request processing performance
func BenchmarkElevator_Request(b *testing.B) {
	elev, err := elevator.New("BenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		from := i % 90
		to := from + 10
		direction := domain.DirectionUp
		if i%2 == 0 {
			direction = domain.DirectionDown
			from, to = to, from
		}

		elev.Request(direction, domain.NewFloor(from), domain.NewFloor(to))
	}
}

// BenchmarkElevator_ConcurrentRequests benchmarks concurrent request handling
func BenchmarkElevator_ConcurrentRequests(b *testing.B) {
	elev, err := elevator.New("ConcurrentBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		for pb.Next() {
			from := counter % 90
			to := from + 10
			direction := domain.DirectionUp
			if counter%2 == 0 {
				direction = domain.DirectionDown
				from, to = to, from
			}

			elev.Request(direction, domain.NewFloor(from), domain.NewFloor(to))
			counter++
		}
	})
}

// BenchmarkElevator_StateOperations benchmarks state access operations
func BenchmarkElevator_StateOperations(b *testing.B) {
	elev, err := elevator.New("StateBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = elev.CurrentFloor()
		_ = elev.CurrentDirection()
		_ = elev.Name()
		_ = elev.Directions()
	}
}

// BenchmarkElevator_IsRequestInRange benchmarks range validation
func BenchmarkElevator_IsRequestInRange(b *testing.B) {
	elev, err := elevator.New("RangeBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		from := i % 90
		to := from + 10
		_ = elev.IsRequestInRange(domain.NewFloor(from), domain.NewFloor(to))
	}
}

// BenchmarkElevator_DirectionsOperations benchmarks directions manager operations
func BenchmarkElevator_DirectionsOperations(b *testing.B) {
	elev, err := elevator.New("DirectionsBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	// Add some requests to create internal state
	for i := 0; i < 10; i++ {
		elev.Request(domain.DirectionUp, domain.NewFloor(i), domain.NewFloor(i+5))
		elev.Request(domain.DirectionDown, domain.NewFloor(i+10), domain.NewFloor(i+5))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		directions := elev.Directions()
		_ = directions.UpDirectionLength()
		_ = directions.DownDirectionLength()
		_ = directions.DirectionsLength()
	}
}

// BenchmarkElevator_MemoryUsage benchmarks memory usage under load
func BenchmarkElevator_MemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		elev, err := elevator.New("MemoryBenchmarkElevator", 0, 50, time.Millisecond*10, time.Millisecond*10,
			30*time.Second, 5, 30*time.Second, 3, 12)
		if err != nil {
			b.Fatal(err)
		}

		// Add multiple requests to simulate real usage
		for j := 0; j < 10; j++ {
			elev.Request(domain.DirectionUp, domain.NewFloor(j), domain.NewFloor(j+5))
			elev.Request(domain.DirectionDown, domain.NewFloor(j+10), domain.NewFloor(j+5))
		}

		// Access various properties
		_ = elev.CurrentFloor()
		_ = elev.CurrentDirection()
		_ = elev.Directions()

		// Properly cleanup the elevator to prevent goroutine leaks
		elev.Shutdown()
	}
}

// BenchmarkElevator_ConcurrentStateAccess benchmarks concurrent state access
func BenchmarkElevator_ConcurrentStateAccess(b *testing.B) {
	elev, err := elevator.New("ConcurrentStateBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	// Add some initial requests
	for i := 0; i < 10; i++ {
		elev.Request(domain.DirectionUp, domain.NewFloor(i), domain.NewFloor(i+10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent reads of elevator state
			_ = elev.CurrentFloor()
			_ = elev.CurrentDirection()
			_ = elev.Name()
			directions := elev.Directions()
			_ = directions.UpDirectionLength()
			_ = directions.DownDirectionLength()
		}
	})
}

// BenchmarkElevator_StatusOperations benchmarks status and health metrics
func BenchmarkElevator_StatusOperations(b *testing.B) {
	elev, err := elevator.New("StatusBenchmarkElevator", 0, 100, time.Millisecond*10, time.Millisecond*10,
		30*time.Second, 5, 30*time.Second, 3, 12)
	if err != nil {
		b.Fatal(err)
	}
	defer elev.Shutdown() // Ensure cleanup

	// Add some requests to create realistic state
	for i := 0; i < 5; i++ {
		elev.Request(domain.DirectionUp, domain.NewFloor(i), domain.NewFloor(i+10))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = elev.GetStatus()
		_ = elev.GetHealthMetrics()
		_ = elev.MinFloor()
		_ = elev.MaxFloor()
	}
}
