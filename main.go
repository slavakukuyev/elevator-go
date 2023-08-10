package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func initLogger() {
	config := zap.Config{
		Encoding:    "console", // or "json"
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			LevelKey:    "level",
			TimeKey:     "time",
			MessageKey:  "message",
			EncodeLevel: zapcore.LowercaseColorLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flushes buffer, if any
}

const _directionUp = "up"
const _directionDown = "down"

// ***************************************************************************

type Destinations struct {
	up   map[int][]int
	down map[int][]int
	//mu   sync.RWMutex
}

func NewDestinations() *Destinations {
	return &Destinations{
		up:   make(map[int][]int),
		down: make(map[int][]int),
	}
}

func (d *Destinations) Append(direction string, fromFloor, toFloor int) {
	// d.mu.Lock()
	// defer d.mu.Unlock()

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

func (d *Destinations) Flush(direction string, fromFloor int) {
	// d.mu.Lock()
	// defer d.mu.Unlock()

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

func (d *Destinations) isUpExisting() bool {
	//d.mu.RLock()
	existing := len(d.up) > 0
	//d.mu.RUnlock()
	return existing
}

func (d *Destinations) isDownExisting() bool {
	//d.mu.RLock()
	existing := len(d.down) > 0
	//d.mu.RUnlock()
	return existing
}

//***************************************************************************

type Elevator struct {
	name         string
	maxFloor     int
	minFloor     int
	currentFloor int
	direction    string
	mu           sync.RWMutex
	destinations *Destinations
	switchOnChan chan byte // Channel for status updates
}

func NewElevator(name string) *Elevator {
	e := &Elevator{
		name:         name,
		maxFloor:     9,
		minFloor:     0,
		currentFloor: 0,
		destinations: NewDestinations(),
		switchOnChan: make(chan byte, 10),
	}

	go e.switchOn()
	return e
}

func (e *Elevator) switchOn() {
	for range e.switchOnChan {
		if e.destinations.isUpExisting() || e.destinations.isDownExisting() {
			e.Run()
		}
	}
}

func (e *Elevator) Run() {
	e.mu.Lock()
	defer e.mu.Unlock()

	logger.Debug("current floor", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
	time.Sleep(time.Millisecond * 500)

	if e.direction == _directionUp && e.destinations.isUpExisting() {
		if _, exists := e.destinations.up[e.currentFloor]; exists {
			e.openDoor()
			e.destinations.Flush(e.direction, e.currentFloor)
			e.closeDoor()
		}

		//if elevator arrived to the top
		if e.currentFloor == e.maxFloor {
			e.direction = _directionDown
			return
		}

		if e.shouldMoveUp() {
			e.currentFloor++
			e.switchOnChan <- 1
			return
		}

	}
	//direction down && requests are down
	if e.direction == _directionDown && e.destinations.isDownExisting() {
		if _, exists := e.destinations.down[e.currentFloor]; exists {
			e.openDoor()
			e.destinations.Flush(e.direction, e.currentFloor)
			e.closeDoor()
		}

		//check if elevator arrived to the bottom
		if e.currentFloor == e.minFloor {
			e.direction = _directionUp
			return
		}

		if e.shouldMoveDown() {
			e.currentFloor--
			e.switchOnChan <- 1
			return
		}

	}

	//case of elevator moving down && no more requests to move down BUT there is a request to move up on the smallest floor
	//smallest floor of the UP direction which is smaller then current floor
	if e.direction == _directionDown && e.destinations.isUpExisting() {
		smallest := findSmallestKey(e.destinations.up)
		if smallest < e.currentFloor {
			e.currentFloor--
			e.switchOnChan <- 1
			return
		}

		if smallest == e.currentFloor {
			e.direction = _directionUp
			e.switchOnChan <- 1
			return
		}
	}

	//the edge case when elevator moving up && there is no more requests to move up BUT new requests are existing to movedown from the largest floor
	//largest floor of the DOWN direction which is greater then current floor
	if e.direction == _directionUp && e.destinations.isDownExisting() {
		largest := findLargestKey(e.destinations.down)
		if largest > e.currentFloor {
			e.currentFloor++
			e.switchOnChan <- 1
			return
		}

		if largest == e.currentFloor {
			e.direction = _directionDown
			e.switchOnChan <- 1
			return
		}
	}

}

// check if elevator has more requests in the up direction and should continue move up
func (e *Elevator) shouldMoveUp() bool {
	if e.destinations.isUpExisting() {
		largest := findLargestKey(e.destinations.up)
		return largest > e.currentFloor
	}

	return false
}

// check if elevator has more requests in the down direction and should continue move down
func (e *Elevator) shouldMoveDown() bool {
	if e.destinations.isDownExisting() {
		smallest := findSmallestKey(e.destinations.down)
		return smallest < e.currentFloor
	}

	return false
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

func (e *Elevator) openDoor() {
	logger.Info("Open doors", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
	time.Sleep(time.Second * 2)
}

func (e *Elevator) closeDoor() {
	logger.Info("Close doors", zap.String("elevator", e.name), zap.Int("floor", e.currentFloor))
}

// Append the request to the elevator regardless of the current direction if `must` is `true`.
// If the direction is empty, set the requested direction.
// If the current direction is opposite, reject the request.
func (e *Elevator) Request(direction string, fromFloor, toFloor int, must bool) bool {
	currentDirection := e.GetDirection()

	if !must && currentDirection != "" && currentDirection != direction {
		return false
	}

	if currentDirection == "" {
		e.SetDirection(direction)
	}

	e.destinations.Append(direction, fromFloor, toFloor)
	e.switchOnChan <- 1
	return true
}

func (e *Elevator) GetDirection() string {
	e.mu.RLock()
	direction := e.direction
	e.mu.RUnlock()
	return direction
}

func (e *Elevator) SetDirection(direction string) {
	e.mu.Lock()
	e.direction = direction
	e.mu.Unlock()
}

// ***********************************************************************************************************
type Manager struct {
	mu        sync.RWMutex
	elevators []*Elevator
}

func NewManager() *Manager {
	return &Manager{
		elevators: make([]*Elevator, 0),
	}
}

func (m *Manager) AddElevator(elevator *Elevator) {
	m.elevators = append(m.elevators, elevator)
}

func (m *Manager) RequestElevator(fromFloor, toFloor int) error {

	if toFloor == fromFloor {
		return fmt.Errorf("%d floor is equal to the same %d floor. The requested floor should be different from your floor", toFloor, fromFloor)
	}

	direction := _directionUp
	if toFloor < fromFloor {
		direction = _directionDown
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	approved := false
	for _, elevator := range m.elevators {
		approved = elevator.Request(direction, fromFloor, toFloor, false)
	}

	if !approved {
		m.elevators[0].Request(direction, fromFloor, toFloor, true)
	}

	return nil

}

func main() {

	var wg sync.WaitGroup
	wg.Add(1)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start your main logic in a goroutine
	go func() {
		defer wg.Done()
		for {
			select {
			case <-signals:
				logger.Info("received termination signal.")
				return // Exit the loop when a termination signal is received
			default:
				time.Sleep(time.Second * 5)
			}
		}
	}()

	initLogger()

	manager := NewManager()

	elevator1 := NewElevator("E1")
	// elevator2 := NewElevator()

	manager.AddElevator(elevator1)
	// manager.AddElevator(elevator2)

	// Request an elevator going from floor 1 to floor 9
	if err := manager.RequestElevator(1, 9); err != nil {
		logger.Error("request elevator 1,9 error", zap.Error(err))
	}

	// Request an elevator going from floor 3 to floor 5
	if err := manager.RequestElevator(3, 5); err != nil {
		logger.Error("request elevator 3,5 error", zap.Error(err))
	}

	// Request an elevator going from floor 3 to floor 5
	if err := manager.RequestElevator(6, 4); err != nil {
		logger.Error("request elevator 6,4 error", zap.Error(err))
	}

	time.Sleep(time.Second * 7)

	if err := manager.RequestElevator(1, 2); err != nil {
		logger.Error("request elevator 1,2 error", zap.Error(err))
	}

	time.Sleep(time.Second * 10)

	if err := manager.RequestElevator(7, 0); err != nil {
		logger.Error("request elevator 7,0 error", zap.Error(err))
	}

	wg.Wait() // Wait until the termination
}
