package main

import "sync"

type Directions struct {
	up   map[int][]int
	down map[int][]int
	mu   sync.RWMutex
}

func NewDirections() *Directions {
	return &Directions{
		up:   make(map[int][]int),
		down: make(map[int][]int),
	}
}

func (d *Directions) Append(direction string, fromFloor, toFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == _directionUp {
		d.up[fromFloor] = append(d.up[fromFloor], toFloor)
		return
	}

	if direction == _directionDown {
		d.down[fromFloor] = append(d.down[fromFloor], toFloor)
	}
}

//creates new keys in the same direction
//removes the request from current floor
/** example:
step 1 : map[1] = [3,5] // from the 1st floor requested floors 3 and 5
step 2 : map[3] = []; map[5] = []; delete map[1] // Elevalor arrived to 1st floor, requested for himself 3,5 floors, and removed 1st floor from the direction slice
step 3: delete map[3] // elevator arrived to 3d floor
step 4: delete map[5] // elevator arrived to 5th floor

the same steps in the opposite direction
*/

func (d *Directions) Flush(direction string, fromFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == _directionUp {
		if len(d.up[fromFloor]) > 0 {
			for _, floor := range d.up[fromFloor] {
				if _, exists := d.up[floor]; !exists {
					d.up[floor] = make([]int, 0)
				}
			}

		}

		delete(d.up, fromFloor)
	}

	if direction == _directionDown {
		if len(d.down[fromFloor]) > 0 {
			for _, floor := range d.down[fromFloor] {
				if _, exists := d.down[floor]; !exists {
					d.down[floor] = make([]int, 0)
				}
			}

		}

		delete(d.down, fromFloor)
	}

}

func (d *Directions) isUpExisting() bool {
	d.mu.RLock()
	existing := len(d.up) > 0
	d.mu.RUnlock()
	return existing
}

func (d *Directions) isDownExisting() bool {
	d.mu.RLock()
	existing := len(d.down) > 0
	d.mu.RUnlock()
	return existing
}

func findLargestKey(m map[int][]int) int {
	largest := 0

	for key := range m {
		if key > largest {
			largest = key
		}
	}

	return largest
}

func findSmallestKey(m map[int][]int) int {
	smallest := 0
	first := true

	for key := range m {
		if first || key < smallest {
			smallest = key
			first = false
		}
	}

	return smallest
}