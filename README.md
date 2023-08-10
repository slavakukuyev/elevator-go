# elevator-go
Golang elevator real system implementation



* create env for elevators values
* create system on each floor: which elevator will stop
* add logic: when first request is arrived:  the lift with  less distance to the floor will move.
* add GitHub Actions
* add Logic of floor under 0 (parking)
* use zap logger


current flow test:
parameters:
```go
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

	time.Sleep(time.Second * 10)

	if err := manager.RequestElevator(1, 2); err != nil {
		logger.Error("request elevator 1,2 error", zap.Error(err))
	}

	time.Sleep(time.Second * 15)

	if err := manager.RequestElevator(7, 0); err != nil {
		logger.Error("request elevator 7,0 error", zap.Error(err))
	}
```

output:
```bash
The elevator E1 is on the 0 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
Elevator E1 opened the doors at floor 5
Elevator E1 closed the doors at floor 5
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 9 floor
Elevator E1 opened the doors at floor 9
Elevator E1 closed the doors at floor 9
The elevator E1 is on the 9 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
Received termination signal.
PS C:\repos\elevator-go> go run .\main.go
The elevator E1 is on the 0 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
Elevator E1 opened the doors at floor 5
Elevator E1 closed the doors at floor 5
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 9 floor
Elevator E1 opened the doors at floor 9
Elevator E1 closed the doors at floor 9
The elevator E1 is on the 9 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 6 floor
Elevator E1 opened the doors at floor 6
Elevator E1 closed the doors at floor 6
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
Elevator E1 opened the doors at floor 4
Elevator E1 closed the doors at floor 4
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
Received termination signal.
PS C:\repos\elevator-go> go run .\main.go
The elevator E1 is on the 0 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
Elevator E1 opened the doors at floor 5
Elevator E1 closed the doors at floor 5
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 9 floor
Elevator E1 opened the doors at floor 9
Elevator E1 closed the doors at floor 9
The elevator E1 is on the 9 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 6 floor
Elevator E1 opened the doors at floor 6
Elevator E1 closed the doors at floor 6
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
Elevator E1 opened the doors at floor 4
Elevator E1 closed the doors at floor 4
The elevator E1 is on the 4 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 4 floor
Received termination signal.
PS C:\repos\elevator-go> go run .\main.go
The elevator E1 is on the 0 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
Elevator E1 opened the doors at floor 5
Elevator E1 closed the doors at floor 5
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 9 floor
Elevator E1 opened the doors at floor 9
Elevator E1 closed the doors at floor 9
The elevator E1 is on the 9 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 6 floor
Elevator E1 opened the doors at floor 6
Elevator E1 closed the doors at floor 6
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
Elevator E1 opened the doors at floor 4
Elevator E1 closed the doors at floor 4
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
Elevator E1 opened the doors at floor 2
Elevator E1 closed the doors at floor 2
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 7 floor
Elevator E1 opened the doors at floor 7
Elevator E1 closed the doors at floor 7
The elevator E1 is on the 6 floor
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 0 floor
Elevator E1 opened the doors at floor 0
Elevator E1 closed the doors at floor 0
Received termination signal.
PS C:\repos\elevator-go> go run .\main.go
The elevator E1 is on the 0 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
Elevator E1 opened the doors at floor 3
Elevator E1 closed the doors at floor 3
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
Elevator E1 opened the doors at floor 5
Elevator E1 closed the doors at floor 5
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 9 floor
Elevator E1 opened the doors at floor 9
Elevator E1 closed the doors at floor 9
The elevator E1 is on the 9 floor
The elevator E1 is on the 8 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 6 floor
Elevator E1 opened the doors at floor 6
Elevator E1 closed the doors at floor 6
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
Elevator E1 opened the doors at floor 4
Elevator E1 closed the doors at floor 4
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 1 floor
Elevator E1 opened the doors at floor 1
Elevator E1 closed the doors at floor 1
The elevator E1 is on the 2 floor
Elevator E1 opened the doors at floor 2
Elevator E1 closed the doors at floor 2
The elevator E1 is on the 2 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 5 floor
The elevator E1 is on the 6 floor
The elevator E1 is on the 7 floor
The elevator E1 is on the 7 floor
Elevator E1 opened the doors at floor 7
Elevator E1 closed the doors at floor 7
The elevator E1 is on the 6 floor
The elevator E1 is on the 5 floor
The elevator E1 is on the 4 floor
The elevator E1 is on the 3 floor
The elevator E1 is on the 2 floor
The elevator E1 is on the 1 floor
The elevator E1 is on the 0 floor
Elevator E1 opened the doors at floor 0
Elevator E1 closed the doors at floor 0
```