package main

import (
	"testing"
)

// TestFindNearestElevator tests the findNearestElevator function
func TestFindNearestElevator(t *testing.T) {
	// Create some sample elevators
	e1 := &Elevator{name: "1"}
	e2 := &Elevator{name: "2"}
	e3 := &Elevator{name: "3"}

	// Create a struct type to hold each test case
	type testCase struct {
		name           string            // name of the test case
		elevators      map[*Elevator]int // map of elevators and their current floors
		requestedFloor int               // requested floor
		want           *Elevator         // expected elevator
	}

	// Create a slice of test cases
	testCases := []testCase{
		{
			name:           "empty map",
			elevators:      map[*Elevator]int{},
			requestedFloor: 5,
			want:           nil,
		},
		{
			name:           "one elevator",
			elevators:      map[*Elevator]int{e1: 3},
			requestedFloor: 5,
			want:           e1,
		},
		{
			name:           "two elevators with different distances",
			elevators:      map[*Elevator]int{e1: 3, e2: 8},
			requestedFloor: 5,
			want:           e1,
		},

		{
			name:           "two elevators with different distances and negative requested floor",
			elevators:      map[*Elevator]int{e1: -5, e3: 6},
			requestedFloor: -1,
			want:           e1,
		},

		{
			name:           "two elevators with different distances and positive requested floor",
			elevators:      map[*Elevator]int{e1: -5, e3: 6},
			requestedFloor: 2,
			want:           e3,
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := findNearestElevator(tc.elevators, tc.requestedFloor)
			if tc.want == nil {
				if got != nil {
					t.Errorf("findNearestElevator(%v, %d) = %v, want nil", tc.elevators, tc.requestedFloor, got)
				}
			} else {
				if got == nil || got.name != tc.want.name {
					t.Errorf("findNearestElevator(%v, %d) = %v, want %v", tc.elevators, tc.requestedFloor, got, tc.want)
				}
			}
		})
	}

	testCases2 := []testCase{
		{
			name:           "two elevators with the same distances and positive requested floor",
			elevators:      map[*Elevator]int{e1: 3, e2: 7, e3: 10},
			requestedFloor: 5,
		},

		{
			name:           "three elevators with the same distances and one negative floor",
			elevators:      map[*Elevator]int{e1: -2, e2: 0, e3: 10},
			requestedFloor: -1,
		},
	}

	for _, tc := range testCases2 {
		t.Run(tc.name, func(t *testing.T) {
			got := findNearestElevator(tc.elevators, tc.requestedFloor)
			if got == nil {
				t.Errorf("findNearestElevator(%v, %d) = %v, want %v", tc.elevators, tc.requestedFloor, got, "One of elevators")
			}

			if got == e3 {
				t.Errorf("findNearestElevator(%v, %d) = %v, want %v", tc.elevators, tc.requestedFloor, got, e3)
			}
		})
	}
}
