package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	httpPkg "github.com/slavakukuyev/elevator-go/internal/http"
)

// TestElevatorServiceIntegration tests the elevator service running in a Docker container.
// This integration test verifies the complete elevator system functionality in an isolated
// containerized environment, ensuring the service works correctly end-to-end.
func TestElevatorServiceIntegration(t *testing.T) {
	// Skip if running in CI without Docker
	if testing.Short() {
		t.Skip("Skipping testcontainers test in short mode")
	}

	ctx := context.Background()

	// Build and start the elevator service container
	t.Logf("🚀 Starting elevator service container build...")
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../..", // Go up two levels to project root
			Dockerfile: "build/package/Dockerfile",
		},
		ExposedPorts: []string{"6660/tcp"},
		Env: map[string]string{
			"ENV":                    "development", // Use development for easier debugging
			"LOG_LEVEL":              "INFO",        // More logging for debugging
			"PORT":                   "6660",
			"DEFAULT_MIN_FLOOR":      "-5",
			"DEFAULT_MAX_FLOOR":      "25",
			"EACH_FLOOR_DURATION":    "50ms", // Fast for testing
			"OPEN_DOOR_DURATION":     "50ms", // Fast for testing
			"DEFAULT_ELEVATOR_COUNT": "0",    // Start with no elevators
			"METRICS_ENABLED":        "true",
			"HEALTH_ENABLED":         "true",
			"WEBSOCKET_ENABLED":      "false", // Disable for simpler testing
			"CORS_ENABLED":           "true",
		},
		WaitingFor: wait.ForHTTP("/v1/health/live").
			WithPort("6660/tcp").
			WithStartupTimeout(120 * time.Second). // Increased timeout for build + startup
			WithPollInterval(2 * time.Second),
	}

	t.Logf("⏳ Building and starting container (this may take 2-3 minutes)...")
	elevatorContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Logf("❌ Container creation failed: %v", err)
		require.NoError(t, err) // This will fail the test with the original error
	}
	t.Logf("✅ Container started successfully!")
	defer func() {
		if logs, logErr := elevatorContainer.Logs(ctx); logErr == nil {
			t.Logf("Container logs available for debugging")
			_ = logs
		}
		_ = elevatorContainer.Terminate(ctx)
	}()

	// Get container endpoint
	host, err := elevatorContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := elevatorContainer.MappedPort(ctx, "6660")
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	t.Logf("Elevator service running at %s", baseURL)

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Test 1: Health Check
	t.Run("Health Check", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/v1/health/live")
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Logf("✅ Health check passed")
	})

	// Test 2: Metrics Endpoint
	t.Run("Metrics Endpoint", func(t *testing.T) {
		resp, err := client.Get(baseURL + "/metrics")
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Logf("✅ Metrics endpoint accessible")
	})

	// Test 3: Create Elevator
	t.Run("Create Elevator", func(t *testing.T) {
		elevator := httpPkg.ElevatorRequestBody{
			Name:     "IntegrationTestElevator",
			MinFloor: -5,
			MaxFloor: 25,
		}

		jsonBody, err := json.Marshal(elevator)
		require.NoError(t, err)

		resp, err := client.Post(baseURL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		t.Logf("✅ Elevator created successfully")

		// Wait for elevator to be ready
		time.Sleep(100 * time.Millisecond)
	})

	// Test 4: Floor Requests - table-driven test following Go best practices
	t.Run("Floor Requests", func(t *testing.T) {
		testCases := []struct {
			name     string
			from, to int
			expected int
		}{
			{"Ground to upper floor", 0, 10, http.StatusOK},
			{"Upper to ground", 15, 0, http.StatusOK},
			{"Basement to upper", -3, 20, http.StatusOK},
			{"Same floor (should be rejected)", 5, 5, http.StatusBadRequest},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				floorRequest := httpPkg.FloorRequestBody{
					From: tc.from,
					To:   tc.to,
				}

				jsonBody, err := json.Marshal(floorRequest)
				require.NoError(t, err)

				resp, err := client.Post(baseURL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				defer func() {
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}
				}()

				assert.Equal(t, tc.expected, resp.StatusCode)
				t.Logf("✅ Floor request %d→%d: %s", tc.from, tc.to, resp.Status)
			})
		}
	})

	// Test 5: Error Handling - validate input sanitization and proper error responses
	t.Run("Error Handling", func(t *testing.T) {
		t.Run("Invalid floor request", func(t *testing.T) {
			floorRequest := httpPkg.FloorRequestBody{
				From: 100, // Out of range
				To:   0,
			}

			jsonBody, err := json.Marshal(floorRequest)
			require.NoError(t, err)

			resp, err := client.Post(baseURL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			t.Logf("✅ Invalid floor request properly rejected")
		})

		t.Run("Invalid elevator creation", func(t *testing.T) {
			elevator := httpPkg.ElevatorRequestBody{
				Name:     "", // Invalid empty name
				MinFloor: 0,
				MaxFloor: 10,
			}

			jsonBody, err := json.Marshal(elevator)
			require.NoError(t, err)

			resp, err := client.Post(baseURL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			t.Logf("✅ Invalid elevator creation properly rejected")
		})
	})

	// Test 6: Multiple Elevators and System Optimization
	t.Run("Multiple Elevators", func(t *testing.T) {
		// Create additional elevators for load testing
		elevators := []httpPkg.ElevatorRequestBody{
			{Name: "Tower1", MinFloor: 0, MaxFloor: 15},
			{Name: "Tower2", MinFloor: 0, MaxFloor: 25},
		}

		for _, elevator := range elevators {
			jsonBody, err := json.Marshal(elevator)
			require.NoError(t, err)

			resp, err := client.Post(baseURL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		// Wait for elevators to be ready
		time.Sleep(200 * time.Millisecond)

		// Test concurrent requests to validate system resilience
		t.Run("Concurrent Requests", func(t *testing.T) {
			requests := []httpPkg.FloorRequestBody{
				{From: 0, To: 10},
				{From: 5, To: 20},
				{From: 15, To: 0},
				{From: 1, To: 12},
				{From: 8, To: 3},
			}

			results := make(chan error, len(requests))

			for _, req := range requests {
				go func(r httpPkg.FloorRequestBody) {
					jsonBody, err := json.Marshal(r)
					if err != nil {
						results <- fmt.Errorf("marshal error: %w", err)
						return
					}

					resp, err := client.Post(baseURL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
					if err != nil {
						results <- fmt.Errorf("request error: %w", err)
						return
					}
					if err := resp.Body.Close(); err != nil {
						t.Logf("Failed to close response body: %v", err)
					}

					if resp.StatusCode != http.StatusOK {
						results <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
						return
					}
					results <- nil
				}(req)
			}

			// Wait for all requests to complete
			for i := 0; i < len(requests); i++ {
				err := <-results
				assert.NoError(t, err)
			}

			t.Logf("✅ All concurrent requests handled successfully")
		})
	})

	t.Logf("🎉 Integration test completed successfully! Service running at %s", baseURL)
}

// TestContainerizedSystemWorkflow tests a complete end-to-end workflow
// simulating real-world usage patterns in an office building environment.
func TestContainerizedSystemWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive workflow test in short mode")
	}

	ctx := context.Background()

	// Start the elevator service with production-like configuration
	t.Logf("🚀 Starting elevator service container for workflow test...")
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../..",
			Dockerfile: "build/package/Dockerfile",
		},
		ExposedPorts: []string{"6660/tcp"},
		Env: map[string]string{
			"ENV":                    "testing",
			"LOG_LEVEL":              "WARN",
			"PORT":                   "6660",
			"DEFAULT_MIN_FLOOR":      "-2",
			"DEFAULT_MAX_FLOOR":      "30",
			"EACH_FLOOR_DURATION":    "20ms",
			"OPEN_DOOR_DURATION":     "20ms",
			"DEFAULT_ELEVATOR_COUNT": "0",
			"METRICS_ENABLED":        "true",
			"HEALTH_ENABLED":         "true",
		},
		WaitingFor: wait.ForHTTP("/v1/health/live").
			WithPort("6660/tcp").
			WithStartupTimeout(120 * time.Second), // Increased timeout
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = container.Terminate(ctx)
	}()

	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "6660")
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	client := &http.Client{Timeout: 15 * time.Second}

	// Simulate office building scenario with realistic usage patterns
	t.Run("Office Building Simulation", func(t *testing.T) {
		// 1. Set up building with multiple elevators
		elevators := []httpPkg.ElevatorRequestBody{
			{Name: "MainElevator", MinFloor: -2, MaxFloor: 30},
			{Name: "ServiceElevator", MinFloor: -2, MaxFloor: 15},
			{Name: "ExpressElevator", MinFloor: 0, MaxFloor: 30},
		}

		for i, elevator := range elevators {
			jsonBody, err := json.Marshal(elevator)
			require.NoError(t, err)

			resp, err := client.Post(baseURL+"/elevator", "application/json", bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			t.Logf("✅ Created %s (%d/%d)", elevator.Name, i+1, len(elevators))
		}

		// Wait for all elevators to be ready
		time.Sleep(300 * time.Millisecond)

		// 2. Simulate morning rush hour traffic patterns
		t.Run("Morning Rush Hour", func(t *testing.T) {
			rushRequests := []httpPkg.FloorRequestBody{
				{From: 0, To: 5}, // Lobby to office floors
				{From: 0, To: 12},
				{From: 0, To: 18},
				{From: 0, To: 25},
				{From: -2, To: 8},  // Parking to office
				{From: -2, To: 15}, // Basement to office
				{From: 0, To: 3},
				{From: 0, To: 22},
			}

			for i, req := range rushRequests {
				jsonBody, err := json.Marshal(req)
				require.NoError(t, err)

				resp, err := client.Post(baseURL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
				// Accept either success (200) or no available elevator (404)
				assert.Contains(t, []int{http.StatusOK, http.StatusNotFound}, resp.StatusCode)
				status := "accepted"
				if resp.StatusCode == http.StatusNotFound {
					status = "no elevator available"
				}
				t.Logf("✅ Rush request %d/%d: %d→%d (%s)", i+1, len(rushRequests), req.From, req.To, status)

				// Small delay between requests to simulate realistic timing
				time.Sleep(10 * time.Millisecond)
			}
		})

		// 3. Regular business hours traffic
		t.Run("Business Hours Traffic", func(t *testing.T) {
			businessRequests := []httpPkg.FloorRequestBody{
				{From: 8, To: 15}, // Inter-floor movement
				{From: 12, To: 3},
				{From: 20, To: 0}, // Going down for lunch
				{From: 5, To: 25},
				{From: 18, To: -2}, // Going to parking
			}

			for _, req := range businessRequests {
				jsonBody, err := json.Marshal(req)
				require.NoError(t, err)

				resp, err := client.Post(baseURL+"/floor", "application/json", bytes.NewBuffer(jsonBody))
				require.NoError(t, err)
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}

			t.Logf("✅ Business hours traffic handled successfully")
		})

		// 4. Validate system observability after load
		t.Run("System Metrics After Load", func(t *testing.T) {
			resp, err := client.Get(baseURL + "/metrics")
			require.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			t.Logf("✅ System metrics available after load testing")
		})

		// 5. Verify system resilience and health
		t.Run("Health Check After Load", func(t *testing.T) {
			resp, err := client.Get(baseURL + "/v1/health/live")
			require.NoError(t, err)
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Logf("Failed to close response body: %v", err)
				}
			}()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			t.Logf("✅ System healthy after comprehensive testing")
		})
	})

	t.Logf("🏢 Office building simulation completed successfully!")
}

// TestWithTestcontainers demonstrates basic testcontainers usage (kept for reference)
// This serves as a simple example of the testcontainers pattern for educational purposes.
func TestWithTestcontainers(t *testing.T) {
	// Skip if running in CI without Docker
	if testing.Short() {
		t.Skip("Skipping testcontainers example in short mode")
	}

	ctx := context.Background()

	// Example: Start an nginx container to demonstrate the pattern
	req := testcontainers.ContainerRequest{
		Image:        "nginx:alpine",
		ExposedPorts: []string{"80/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("80/tcp").WithStartupTimeout(30 * time.Second),
	}

	nginxContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		_ = nginxContainer.Terminate(ctx)
	}()

	// Get the container endpoint
	host, err := nginxContainer.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := nginxContainer.MappedPort(ctx, "80")
	require.NoError(t, err)

	url := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	// Test that the service is running
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	t.Logf("✅ Testcontainers pattern demonstrated with nginx at %s", url)
}
