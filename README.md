# elevator-go
Golang elevator real system implementation



* create env for elevators values
* create system on each floor: which elevator will stop
* add logic: when first request is arrived:  the lift with  less distance to the floor will move.
* add GitHub Actions
* add Logic of floor under 0 (parking)
* use zap logger (done)
* create docker image
* create manager algorytm on base of all elevators


current flow test:
parameters:
```go
	manager := NewManager()

	elevator1 := NewElevator("A")
	elevator2 := NewElevator("B")

	manager.AddElevator(elevator1)
	manager.AddElevator(elevator2)

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
```

output:
```bash
2023-08-10T17:57:23.992+0300    debug   current floor   {"elevator": "A", "floor": 0}
2023-08-10T17:57:23.992+0300    debug   current floor   {"elevator": "B", "floor": 0}
2023-08-10T17:57:24.519+0300    debug   current floor   {"elevator": "B", "floor": 1}
2023-08-10T17:57:24.519+0300    debug   current floor   {"elevator": "A", "floor": 1}
2023-08-10T17:57:25.024+0300    info    open doors      {"elevator": "A", "floor": 1}
2023-08-10T17:57:25.024+0300    debug   current floor   {"elevator": "B", "floor": 2}
2023-08-10T17:57:25.531+0300    debug   current floor   {"elevator": "B", "floor": 3}
2023-08-10T17:57:26.041+0300    info    open doors      {"elevator": "B", "floor": 3}
2023-08-10T17:57:27.030+0300    info    close doors     {"elevator": "A", "floor": 1}
2023-08-10T17:57:27.030+0300    debug   current floor   {"elevator": "A", "floor": 2}
2023-08-10T17:57:27.543+0300    debug   current floor   {"elevator": "A", "floor": 3}
2023-08-10T17:57:28.047+0300    debug   current floor   {"elevator": "A", "floor": 4}
2023-08-10T17:57:28.047+0300    info    close doors     {"elevator": "B", "floor": 3}
2023-08-10T17:57:28.047+0300    debug   current floor   {"elevator": "B", "floor": 4}
2023-08-10T17:57:28.547+0300    debug   current floor   {"elevator": "B", "floor": 5}
2023-08-10T17:57:28.547+0300    debug   current floor   {"elevator": "A", "floor": 5}
2023-08-10T17:57:29.048+0300    debug   current floor   {"elevator": "A", "floor": 6}
2023-08-10T17:57:29.048+0300    info    open doors      {"elevator": "B", "floor": 5}
2023-08-10T17:57:29.565+0300    debug   current floor   {"elevator": "A", "floor": 7}
2023-08-10T17:57:30.065+0300    debug   current floor   {"elevator": "A", "floor": 8}
2023-08-10T17:57:30.568+0300    debug   current floor   {"elevator": "A", "floor": 9}
2023-08-10T17:57:31.051+0300    info    close doors     {"elevator": "B", "floor": 5}
2023-08-10T17:57:31.083+0300    info    open doors      {"elevator": "A", "floor": 9}
2023-08-10T17:57:33.088+0300    info    close doors     {"elevator": "A", "floor": 9}
2023-08-10T17:57:33.089+0300    debug   current floor   {"elevator": "A", "floor": 9}
2023-08-10T17:57:33.602+0300    debug   current floor   {"elevator": "A", "floor": 8}
2023-08-10T17:57:34.104+0300    debug   current floor   {"elevator": "A", "floor": 7}
2023-08-10T17:57:34.620+0300    debug   current floor   {"elevator": "A", "floor": 6}
2023-08-10T17:57:35.136+0300    info    open doors      {"elevator": "A", "floor": 6}
2023-08-10T17:57:37.151+0300    info    close doors     {"elevator": "A", "floor": 6}
2023-08-10T17:57:37.152+0300    debug   current floor   {"elevator": "A", "floor": 5}
2023-08-10T17:57:37.670+0300    debug   current floor   {"elevator": "A", "floor": 4}
2023-08-10T17:57:38.185+0300    info    open doors      {"elevator": "A", "floor": 4}
2023-08-10T17:57:40.203+0300    info    close doors     {"elevator": "A", "floor": 4}
2023-08-10T17:57:40.203+0300    debug   current floor   {"elevator": "A", "floor": 3}
2023-08-10T17:57:40.719+0300    debug   current floor   {"elevator": "A", "floor": 2}
2023-08-10T17:57:41.221+0300    debug   current floor   {"elevator": "A", "floor": 1}
2023-08-10T17:57:41.739+0300    debug   current floor   {"elevator": "A", "floor": 1}
2023-08-10T17:57:42.255+0300    info    open doors      {"elevator": "A", "floor": 1}
2023-08-10T17:57:44.257+0300    info    close doors     {"elevator": "A", "floor": 1}
2023-08-10T17:57:44.259+0300    debug   current floor   {"elevator": "A", "floor": 2}
2023-08-10T17:57:44.775+0300    info    open doors      {"elevator": "A", "floor": 2}
2023-08-10T17:57:46.785+0300    info    close doors     {"elevator": "A", "floor": 2}
2023-08-10T17:57:46.786+0300    debug   current floor   {"elevator": "A", "floor": 2}
2023-08-10T17:57:47.293+0300    debug   current floor   {"elevator": "A", "floor": 3}
2023-08-10T17:57:47.806+0300    debug   current floor   {"elevator": "A", "floor": 4}
2023-08-10T17:57:48.325+0300    debug   current floor   {"elevator": "A", "floor": 5}
2023-08-10T17:57:48.839+0300    debug   current floor   {"elevator": "A", "floor": 6}
2023-08-10T17:57:49.356+0300    debug   current floor   {"elevator": "A", "floor": 7}
2023-08-10T17:57:49.857+0300    debug   current floor   {"elevator": "A", "floor": 7}
2023-08-10T17:57:50.372+0300    info    open doors      {"elevator": "A", "floor": 7}
2023-08-10T17:57:52.377+0300    info    close doors     {"elevator": "A", "floor": 7}
2023-08-10T17:57:52.380+0300    debug   current floor   {"elevator": "A", "floor": 6}
2023-08-10T17:57:52.890+0300    debug   current floor   {"elevator": "A", "floor": 5}
2023-08-10T17:57:53.393+0300    debug   current floor   {"elevator": "A", "floor": 4}
2023-08-10T17:57:53.895+0300    debug   current floor   {"elevator": "A", "floor": 3}
2023-08-10T17:57:54.411+0300    debug   current floor   {"elevator": "A", "floor": 2}
2023-08-10T17:57:54.925+0300    debug   current floor   {"elevator": "A", "floor": 1}
2023-08-10T17:57:55.431+0300    debug   current floor   {"elevator": "A", "floor": 0}
2023-08-10T17:57:55.945+0300    info    open doors      {"elevator": "A", "floor": 0}
2023-08-10T17:57:57.947+0300    info    close doors     {"elevator": "A", "floor": 0}
```
