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

func (d *Directions) UpDirectionLength() int {
	d.mu.RLock()
	l := len(d.up)
	d.mu.RUnlock()
	return l
}

func (d *Directions) DownDirectionLength() int {
	d.mu.RLock()
	l := len(d.down)
	d.mu.RUnlock()
	return l
}

func (d *Directions) DirectionsLength() int {
	d.mu.RLock()
	l := len(d.up) + len(d.down)
	d.mu.RUnlock()
	return l
}

func (d *Directions) IsExisting(direction string, from, to int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if direction == _directionUp && validateIntInMapSlice(d.up, from, to) {
		return true
	}

	if direction == _directionDown && validateIntInMapSlice(d.down, from, to) {
		return true
	}

	return false
}

func validateIntInMapSlice(m map[int][]int, key, value int) bool {
	slice, exists := m[key]
	if !exists {
		return false
	}

	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}
