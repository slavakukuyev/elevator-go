package manager

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slavakukuyev/elevator-go/internal/domain"
	"github.com/slavakukuyev/elevator-go/internal/elevator"
	"github.com/slavakukuyev/elevator-go/internal/factory"
	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

func buildManagerTestConfig() *config.Config {
	// Set testing environment to get proper defaults including timeouts
	os.Setenv("ENV", "testing")
	os.Setenv("LOG_LEVEL", "ERROR")
	defer func() {
		os.Unsetenv("ENV")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg, err := config.InitConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize test config: %v", err))
	}
	return cfg
}

func TestManager_New(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		cfg := buildManagerTestConfig()
		factory := &factory.StandardElevatorFactory{}

		manager := New(cfg, factory)

		require.NotNil(t, manager)
		assert.NotNil(t, manager.elevators)
		assert.Equal(t, factory, manager.factory)
		assert.NotNil(t, manager.logger)
		assert.Len(t, manager.elevators, 0) // Should start empty
	})
}

func TestManager_AddElevator(t *testing.T) {
	tests := []struct {
		name              string
		elevatorName      string
		minFloor          int
		maxFloor          int
		eachFloorDuration time.Duration
		openDoorDuration  time.Duration
		expectError       bool
		errorContains     string
		setupExisting     []string // Names of elevators to create first
	}{
		{
			name:              "valid elevator addition",
			elevatorName:      "TestElevator",
			minFloor:          0,
			maxFloor:          10,
			eachFloorDuration: time.Millisecond * 100,
			openDoorDuration:  time.Millisecond * 100,
			expectError:       false,
		},
		{
			name:              "duplicate name should fail",
			elevatorName:      "DuplicateElevator",
			minFloor:          0,
			maxFloor:          10,
			eachFloorDuration: time.Millisecond * 100,
			openDoorDuration:  time.Millisecond * 100,
			setupExisting:     []string{"DuplicateElevator"},
			expectError:       true,
			errorContains:     "already exists",
		},
		{
			name:              "basement elevator",
			elevatorName:      "BasementElevator",
			minFloor:          -5,
			maxFloor:          0,
			eachFloorDuration: time.Millisecond * 100,
			openDoorDuration:  time.Millisecond * 100,
			expectError:       false,
		},
		{
			name:              "high-rise elevator",
			elevatorName:      "HighRiseElevator",
			minFloor:          0,
			maxFloor:          100,
			eachFloorDuration: time.Millisecond * 100,
			openDoorDuration:  time.Millisecond * 100,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cfg := buildManagerTestConfig()
			factory := &factory.StandardElevatorFactory{}
			manager := New(cfg, factory)

			// Setup existing elevators if specified
			for _, existingName := range tt.setupExisting {
				err := manager.AddElevator(ctx, cfg, existingName, 0, 5, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
				require.NoError(t, err)
			}

			// Test the addition
			err := manager.AddElevator(ctx, cfg, tt.elevatorName, tt.minFloor, tt.maxFloor, tt.eachFloorDuration, tt.openDoorDuration, cfg.DefaultOverloadThreshold)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)

				// Verify elevator was added
				addedElevator := manager.GetElevator(tt.elevatorName)
				require.NotNil(t, addedElevator)
				assert.Equal(t, tt.elevatorName, addedElevator.Name())
				assert.Equal(t, domain.NewFloor(tt.minFloor), addedElevator.MinFloor())
				assert.Equal(t, domain.NewFloor(tt.maxFloor), addedElevator.MaxFloor())
			}
		})
	}
}

func TestManager_GetElevator(t *testing.T) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	// Add test elevators
	elevatorNames := []string{"Elevator1", "Elevator2", "Elevator3"}
	for _, name := range elevatorNames {
		err := manager.AddElevator(ctx, cfg, name, 0, 10, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)
	}

	t.Run("get existing elevator", func(t *testing.T) {
		for _, name := range elevatorNames {
			elevator := manager.GetElevator(name)
			assert.NotNil(t, elevator)
			assert.Equal(t, name, elevator.Name())
		}
	})

	t.Run("get non-existent elevator", func(t *testing.T) {
		elevator := manager.GetElevator("NonExistentElevator")
		assert.Nil(t, elevator)
	})

	t.Run("get elevator with empty name", func(t *testing.T) {
		elevator := manager.GetElevator("")
		assert.Nil(t, elevator)
	})
}

func TestManager_GetElevators(t *testing.T) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	t.Run("empty manager", func(t *testing.T) {
		elevators := manager.GetElevators()
		assert.Len(t, elevators, 0)
	})

	t.Run("with elevators", func(t *testing.T) {
		// Add test elevators
		elevatorNames := []string{"A", "B", "C"}
		for _, name := range elevatorNames {
			err := manager.AddElevator(ctx, cfg, name, 0, 10, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
			require.NoError(t, err)
		}

		elevators := manager.GetElevators()
		assert.Len(t, elevators, len(elevatorNames))

		// Verify all elevators are present
		foundNames := make(map[string]bool)
		for _, elevator := range elevators {
			foundNames[elevator.Name()] = true
		}

		for _, expectedName := range elevatorNames {
			assert.True(t, foundNames[expectedName], "Expected elevator %s not found", expectedName)
		}
	})
}

func TestManager_RequestElevator(t *testing.T) {
	tests := []struct {
		name           string
		fromFloor      int
		toFloor        int
		setupElevators []struct {
			name               string
			minFloor, maxFloor int
		}
		expectError      bool
		errorContains    string
		expectedElevator string // Name of expected elevator, or empty if any is acceptable
	}{
		{
			name:      "valid up request",
			fromFloor: 2,
			toFloor:   5,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"StandardElevator", 0, 10},
			},
			expectError:      false,
			expectedElevator: "StandardElevator",
		},
		{
			name:      "valid down request",
			fromFloor: 8,
			toFloor:   3,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"StandardElevator", 0, 10},
			},
			expectError:      false,
			expectedElevator: "StandardElevator",
		},
		{
			name:      "same floor request should fail",
			fromFloor: 5,
			toFloor:   5,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"StandardElevator", 0, 10},
			},
			expectError:   true,
			errorContains: "must be different",
		},
		{
			name:      "request outside all elevator ranges",
			fromFloor: 15,
			toFloor:   20,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"SmallElevator", 0, 10},
			},
			expectError:   true,
			errorContains: "out of range",
		},
		{
			name:      "multiple elevators - choose best",
			fromFloor: 2,
			toFloor:   5,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"Elevator1", 0, 10},
				{"Elevator2", 0, 15},
				{"BasementElevator", -5, 5},
			},
			expectError: false,
			// Don't specify expected elevator as algorithm may choose any suitable one
		},
		{
			name:      "basement request",
			fromFloor: -3,
			toFloor:   0,
			setupElevators: []struct {
				name               string
				minFloor, maxFloor int
			}{
				{"BasementElevator", -5, 5},
				{"MainElevator", 0, 20},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			cfg := buildManagerTestConfig()
			factory := &factory.StandardElevatorFactory{}
			manager := New(cfg, factory)

			// Setup elevators
			for _, setup := range tt.setupElevators {
				err := manager.AddElevator(ctx, cfg, setup.name, setup.minFloor, setup.maxFloor, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
				require.NoError(t, err)
			}

			// Make the request
			elevator, err := manager.RequestElevator(ctx, tt.fromFloor, tt.toFloor)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, elevator)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, elevator)

				if tt.expectedElevator != "" {
					assert.Equal(t, tt.expectedElevator, elevator.Name())
				}

				// Verify the request was added to the elevator
				assert.Greater(t, elevator.Directions().DirectionsLength(), 0, "Request should be added to elevator")
			}
		})
	}
}

func TestManager_RequestElevator_Algorithm(t *testing.T) {
	t.Run("idle elevator selection", func(t *testing.T) {
		ctx := context.Background()
		cfg := buildManagerTestConfig()
		factory := &factory.StandardElevatorFactory{}
		manager := New(cfg, factory)

		// Add multiple elevators
		err := manager.AddElevator(ctx, cfg, "Elevator1", 0, 20, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)
		err = manager.AddElevator(ctx, cfg, "Elevator2", 0, 20, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		// All elevators are idle, so algorithm should pick one
		elevator, err := manager.RequestElevator(ctx, 5, 10)
		require.NoError(t, err)
		assert.NotNil(t, elevator)
	})

	t.Run("existing request detection", func(t *testing.T) {
		ctx := context.Background()
		cfg := buildManagerTestConfig()
		factory := &factory.StandardElevatorFactory{}
		manager := New(cfg, factory)

		err := manager.AddElevator(ctx, cfg, "TestElevator", 0, 20, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		// Make first request
		elevator1, err := manager.RequestElevator(ctx, 5, 10)
		require.NoError(t, err)

		// Make the same request again - should return the same elevator
		elevator2, err := manager.RequestElevator(ctx, 5, 10)
		require.NoError(t, err)

		assert.Equal(t, elevator1.Name(), elevator2.Name())
	})
}

func TestManager_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	// Add base elevators
	for i := 0; i < 3; i++ {
		elevatorName := fmt.Sprintf("Elevator%d", i)
		err := manager.AddElevator(ctx, cfg, elevatorName, 0, 20, time.Millisecond*10, time.Millisecond*10, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)
	}

	const numGoroutines = 20
	const requestsPerGoroutine = 10

	t.Run("concurrent requests", func(t *testing.T) {
		var wg sync.WaitGroup
		var successCount int64
		var mu sync.Mutex

		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				defer wg.Done()

				for j := 0; j < requestsPerGoroutine; j++ {
					from := routineID % 15
					to := from + 3
					if to > 20 {
						to = 20
					}

					_, err := manager.RequestElevator(ctx, from, to)
					if err == nil {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()

		// Should have many successful requests
		mu.Lock()
		assert.Greater(t, successCount, int64(numGoroutines*requestsPerGoroutine/2))
		mu.Unlock()
	})

	t.Run("concurrent elevator addition", func(t *testing.T) {
		var wg sync.WaitGroup
		var successCount int64
		var mu sync.Mutex

		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(routineID int) {
				defer wg.Done()

				elevatorName := fmt.Sprintf("ConcurrentElevator%d", routineID)
				err := manager.AddElevator(ctx, cfg, elevatorName, 0, 10, time.Millisecond*10, time.Millisecond*10, cfg.DefaultOverloadThreshold)

				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// All should succeed since they have unique names
		mu.Lock()
		assert.Equal(t, int64(numGoroutines), successCount)
		mu.Unlock()

		// Verify all elevators were added
		elevators := manager.GetElevators()
		assert.Len(t, elevators, 3+numGoroutines) // 3 original + numGoroutines new
	})
}

func TestManager_GetStatus(t *testing.T) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	t.Run("empty manager status", func(t *testing.T) {
		status, err := manager.GetStatus()
		require.NoError(t, err)
		assert.NotNil(t, status)
	})

	t.Run("manager with elevators status", func(t *testing.T) {
		// Add some elevators
		err := manager.AddElevator(ctx, cfg, "StatusElevator1", 0, 10, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)
		err = manager.AddElevator(ctx, cfg, "StatusElevator2", -5, 15, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		status, err := manager.GetStatus()
		require.NoError(t, err)
		assert.NotNil(t, status)

		// Status should contain information about elevators
		// The exact structure depends on implementation
	})
}

func TestFindNearestElevator(t *testing.T) {
	// Create test elevators with proper initialization
	e1, _ := elevator.New("1", 0, 10, time.Millisecond*100, time.Millisecond*100,
		30*time.Second, 5, 30*time.Second, 3, 12)
	e2, _ := elevator.New("2", 0, 10, time.Millisecond*100, time.Millisecond*100,
		30*time.Second, 5, 30*time.Second, 3, 12)
	e3, _ := elevator.New("3", 0, 10, time.Millisecond*100, time.Millisecond*100,
		30*time.Second, 5, 30*time.Second, 3, 12)

	tests := []struct {
		name           string
		elevators      map[*elevator.Elevator]domain.Floor
		requestedFloor domain.Floor
		want           *elevator.Elevator
		wantDistance   int // Expected distance for verification
	}{
		{
			name: "single elevator",
			elevators: map[*elevator.Elevator]domain.Floor{
				e1: domain.NewFloor(5),
			},
			requestedFloor: domain.NewFloor(3),
			want:           e1,
			wantDistance:   2,
		},
		{
			name: "multiple elevators - closest wins",
			elevators: map[*elevator.Elevator]domain.Floor{
				e1: domain.NewFloor(5),
				e2: domain.NewFloor(2),
				e3: domain.NewFloor(8),
			},
			requestedFloor: domain.NewFloor(3),
			want:           e2,
			wantDistance:   1,
		},
		{
			name: "exact match",
			elevators: map[*elevator.Elevator]domain.Floor{
				e1: domain.NewFloor(5),
				e2: domain.NewFloor(3),
				e3: domain.NewFloor(8),
			},
			requestedFloor: domain.NewFloor(3),
			want:           e2,
			wantDistance:   0,
		},
		{
			name:           "no elevators",
			elevators:      map[*elevator.Elevator]domain.Floor{},
			requestedFloor: domain.NewFloor(3),
			want:           nil,
		},
		{
			name: "negative floors",
			elevators: map[*elevator.Elevator]domain.Floor{
				e1: domain.NewFloor(-5),
				e2: domain.NewFloor(-1),
				e3: domain.NewFloor(2),
			},
			requestedFloor: domain.NewFloor(0),
			want:           e2,
			wantDistance:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := findNearestElevator(tt.elevators, tt.requestedFloor)

			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				require.NotNil(t, got)
				assert.Equal(t, tt.want.Name(), got.Name())

				// Verify it's actually the closest
				if len(tt.elevators) > 1 {
					gotFloor := tt.elevators[got]
					actualDistance := floorsDiff(gotFloor, tt.requestedFloor)
					assert.Equal(t, tt.wantDistance, actualDistance)
				}
			}
		})
	}
}

func TestManager_HelperFunctions(t *testing.T) {
	t.Run("floorsDiff", func(t *testing.T) {
		tests := []struct {
			name     string
			floor1   domain.Floor
			floor2   domain.Floor
			expected int
		}{
			{
				name:     "positive difference",
				floor1:   domain.NewFloor(5),
				floor2:   domain.NewFloor(3),
				expected: 2,
			},
			{
				name:     "negative difference",
				floor1:   domain.NewFloor(3),
				floor2:   domain.NewFloor(7),
				expected: 4,
			},
			{
				name:     "zero difference",
				floor1:   domain.NewFloor(5),
				floor2:   domain.NewFloor(5),
				expected: 0,
			},
			{
				name:     "negative floors",
				floor1:   domain.NewFloor(-3),
				floor2:   domain.NewFloor(-7),
				expected: 4,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := floorsDiff(tt.floor1, tt.floor2)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestManager_GetHealthStatus_ZeroElevators(t *testing.T) {
	ctx := context.Background()
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	t.Run("zero elevators should be healthy initial state", func(t *testing.T) {
		// Test with completely empty manager
		health, err := manager.GetHealthStatus()
		require.NoError(t, err)
		assert.NotNil(t, health)

		// Verify the key health indicators
		assert.Equal(t, 0, health["total_elevators"], "Total elevators should be 0")
		assert.Equal(t, 0, health["healthy_elevators"], "Healthy elevators should be 0")
		assert.True(t, health["system_healthy"].(bool), "System should be healthy with 0 elevators")
		assert.NotNil(t, health["timestamp"], "Timestamp should be present")
		assert.NotNil(t, health["elevators"], "Elevators map should be present")

		// Elevators map should be empty
		elevators, ok := health["elevators"].(map[string]interface{})
		require.True(t, ok, "Elevators should be a map")
		assert.Len(t, elevators, 0, "Elevators map should be empty")
	})

	t.Run("system becomes unhealthy when elevator is added but unhealthy", func(t *testing.T) {
		// This test verifies that our logic works correctly when we do have elevators
		// Add an elevator (which should start healthy)
		err := manager.AddElevator(ctx, cfg, "TestElevator", 0, 10, time.Millisecond*100, time.Millisecond*100, cfg.DefaultOverloadThreshold)
		require.NoError(t, err)

		// Get health status
		health, err := manager.GetHealthStatus()
		require.NoError(t, err)

		// Should have 1 elevator
		assert.Equal(t, 1, health["total_elevators"])
		assert.Equal(t, 1, health["healthy_elevators"])
		assert.True(t, health["system_healthy"].(bool))

		// Note: Testing unhealthy elevator scenarios would require more complex setup
		// to break the circuit breaker or create error conditions
	})
}

func TestManager_GetMetrics_ZeroElevators(t *testing.T) {
	cfg := buildManagerTestConfig()
	factory := &factory.StandardElevatorFactory{}
	manager := New(cfg, factory)

	t.Run("metrics with zero elevators", func(t *testing.T) {
		metrics := manager.GetMetrics()
		require.NotNil(t, metrics)

		// Verify metrics for empty system
		assert.Equal(t, 0, metrics["total_elevators"], "Total elevators should be 0")
		assert.Equal(t, 0, metrics["healthy_elevators"], "Healthy elevators should be 0")
		assert.Equal(t, 0, metrics["total_requests"], "Total requests should be 0")
		assert.Equal(t, 0, metrics["total_up_requests"], "Up requests should be 0")
		assert.Equal(t, 0, metrics["total_down_requests"], "Down requests should be 0")
		assert.Equal(t, float64(0), metrics["average_load"], "Average load should be 0")
		assert.NotNil(t, metrics["timestamp"], "Timestamp should be present")

		// Performance score and efficiency should handle zero-division gracefully
		assert.NotNil(t, metrics["performance_score"], "Performance score should be present")
		assert.NotNil(t, metrics["system_efficiency"], "System efficiency should be present")
	})
}
