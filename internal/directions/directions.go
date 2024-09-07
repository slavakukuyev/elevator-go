package directions

import (
	"sync"

	"github.com/slavakukuyev/elevator-go/internal/infra/config"
)

type T struct {
	up             map[int][]int
	down           map[int][]int
	mu             sync.RWMutex
	_directionUp   string
	_directionDown string
}

func New(cfg *config.Config) *T {
	return &T{
		up:             make(map[int][]int),
		down:           make(map[int][]int),
		_directionUp:   cfg.DirectionUpKey,
		_directionDown: cfg.DirectionDownKey,
	}
}

func (d *T) Append(direction string, fromFloor, toFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == d._directionUp {
		d.up[fromFloor] = append(d.up[fromFloor], toFloor)
		return
	}

	if direction == d._directionDown {
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

func (d *T) Flush(direction string, fromFloor int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if direction == d._directionUp {
		if len(d.up[fromFloor]) > 0 {
			for _, floor := range d.up[fromFloor] {
				if _, exists := d.up[floor]; !exists {
					d.up[floor] = make([]int, 0)
				}
			}

		}

		delete(d.up, fromFloor)
	}

	if direction == d._directionDown {
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

func (d *T) UpDirectionLength() int {
	d.mu.RLock()
	l := len(d.up)
	d.mu.RUnlock()
	return l
}

func (d *T) DownDirectionLength() int {
	d.mu.RLock()
	l := len(d.down)
	d.mu.RUnlock()
	return l
}

func (d *T) DirectionsLength() int {
	d.mu.RLock()
	l := len(d.up) + len(d.down)
	d.mu.RUnlock()
	return l
}

func (d *T) IsExisting(direction string, from, to int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if direction == d._directionUp && isValueInMapSlice(d.up, from, to) {
		return true
	}

	if direction == d._directionDown && isValueInMapSlice(d.down, from, to) {
		return true
	}

	return false
}

func isValueInMapSlice(m map[int][]int, key, value int) bool {
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

func (d *T) Up() map[int][]int {
	return d.up
}

func (d *T) Down() map[int][]int {
	return d.down
}
