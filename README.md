# elevator-go
Golang elevator real system implementation



* create env for elevators values
* add GitHub Actions
* add Logic of floor under 0 (parking)
* use zap logger (done)
* create docker image
* create manager algorytm on base of all elevators (done)
* support min, max for each elevator separately. manager has to handle it


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

	time.Sleep(time.Second * 7)

	if err := manager.RequestElevator(1, 2); err != nil {
		logger.Error("request elevator 1,2 error", zap.Error(err))
	}

	if err := manager.RequestElevator(7, 9); err != nil {
		logger.Error("request elevator 1,2 error", zap.Error(err))
	}

	time.Sleep(time.Second * 10)

	if err := manager.RequestElevator(7, 0); err != nil {
		logger.Error("request elevator 7,0 error", zap.Error(err))
	}

	if err := manager.RequestElevator(3, 0); err != nil {
		logger.Error("request elevator 7,0 error", zap.Error(err))
	}
```

output:
```bash
2023-08-12T17:09:42.272+0300	info	Request has been approved	{"elevator": "A", "fromFloor": 1, "toFloor": 9}
2023-08-12T17:09:42.272+0300	debug	current floor	{"elevator": "A", "floor": 0}
2023-08-12T17:09:42.792+0300	info	Request has been approved	{"elevator": "B", "fromFloor": 3, "toFloor": 5}
2023-08-12T17:09:42.792+0300	debug	current floor	{"elevator": "B", "floor": 0}
2023-08-12T17:09:42.792+0300	debug	current floor	{"elevator": "A", "floor": 1}
2023-08-12T17:09:43.307+0300	info	open doors	{"elevator": "A", "floor": 1}
2023-08-12T17:09:43.307+0300	debug	current floor	{"elevator": "B", "floor": 1}
2023-08-12T17:09:43.819+0300	debug	current floor	{"elevator": "B", "floor": 2}
2023-08-12T17:09:44.327+0300	debug	current floor	{"elevator": "B", "floor": 3}
2023-08-12T17:09:44.839+0300	info	open doors	{"elevator": "B", "floor": 3}
2023-08-12T17:09:45.318+0300	info	close doors	{"elevator": "A", "floor": 1}
2023-08-12T17:09:45.318+0300	debug	current floor	{"elevator": "A", "floor": 2}
2023-08-12T17:09:45.833+0300	debug	current floor	{"elevator": "A", "floor": 3}
2023-08-12T17:09:46.342+0300	debug	current floor	{"elevator": "A", "floor": 4}
2023-08-12T17:09:46.843+0300	debug	current floor	{"elevator": "A", "floor": 5}
2023-08-12T17:09:46.843+0300	info	close doors	{"elevator": "B", "floor": 3}
2023-08-12T17:09:59.474+0300	debug	current floor	{"elevator": "A", "floor": 6}
2023-08-12T17:09:59.474+0300	debug	current floor	{"elevator": "B", "floor": 4}
2023-08-12T17:10:04.720+0300	debug	current floor	{"elevator": "A", "floor": 7}
2023-08-12T17:10:04.720+0300	debug	current floor	{"elevator": "B", "floor": 5}
2023-08-12T17:10:05.235+0300	info	open doors	{"elevator": "B", "floor": 5}
2023-08-12T17:10:05.235+0300	info	Request has been approved	{"elevator": "A", "fromFloor": 6, "toFloor": 4}
2023-08-12T17:10:05.235+0300	debug	current floor	{"elevator": "A", "floor": 8}
2023-08-12T17:10:05.750+0300	debug	current floor	{"elevator": "A", "floor": 9}
2023-08-12T17:10:06.266+0300	info	open doors	{"elevator": "A", "floor": 9}
2023-08-12T17:10:07.242+0300	info	close doors	{"elevator": "B", "floor": 5}
2023-08-12T17:10:08.268+0300	info	close doors	{"elevator": "A", "floor": 9}
2023-08-12T17:10:08.268+0300	debug	current floor	{"elevator": "A", "floor": 9}
2023-08-12T17:10:08.774+0300	debug	current floor	{"elevator": "A", "floor": 8}
2023-08-12T17:10:09.285+0300	debug	current floor	{"elevator": "A", "floor": 7}
2023-08-12T17:10:09.787+0300	debug	current floor	{"elevator": "A", "floor": 6}
2023-08-12T17:10:10.290+0300	info	open doors	{"elevator": "A", "floor": 6}
2023-08-12T17:10:12.305+0300	info	close doors	{"elevator": "A", "floor": 6}
2023-08-12T17:10:37.194+0300	debug	current floor	{"elevator": "A", "floor": 5}
2023-08-12T17:10:37.704+0300	info	Request has been approved	{"elevator": "A", "fromFloor": 1, "toFloor": 2}
2023-08-12T17:10:37.704+0300	debug	current floor	{"elevator": "A", "floor": 4}
2023-08-12T17:10:38.216+0300	info	open doors	{"elevator": "A", "floor": 4}
2023-08-12T17:10:40.231+0300	info	close doors	{"elevator": "A", "floor": 4}
2023-08-12T17:10:40.231+0300	info	Request has been approved	{"elevator": "B", "fromFloor": 7, "toFloor": 9}
2023-08-12T17:10:40.231+0300	debug	current floor	{"elevator": "B", "floor": 5}
2023-08-12T17:10:40.231+0300	debug	current floor	{"elevator": "A", "floor": 3}
2023-08-12T17:10:40.732+0300	debug	current floor	{"elevator": "A", "floor": 2}
2023-08-12T17:10:40.732+0300	debug	current floor	{"elevator": "B", "floor": 6}
2023-08-12T17:10:41.236+0300	debug	current floor	{"elevator": "B", "floor": 7}
2023-08-12T17:10:41.236+0300	debug	current floor	{"elevator": "A", "floor": 1}
2023-08-12T17:10:41.750+0300	debug	current floor	{"elevator": "A", "floor": 1}
2023-08-12T17:10:41.750+0300	info	open doors	{"elevator": "B", "floor": 7}
2023-08-12T17:10:42.261+0300	info	open doors	{"elevator": "A", "floor": 1}
2023-08-12T17:10:43.751+0300	info	close doors	{"elevator": "B", "floor": 7}
2023-08-12T17:10:43.751+0300	debug	current floor	{"elevator": "B", "floor": 8}
2023-08-12T17:10:44.256+0300	debug	current floor	{"elevator": "B", "floor": 9}
2023-08-12T17:10:44.272+0300	info	close doors	{"elevator": "A", "floor": 1}
2023-08-12T17:10:44.272+0300	debug	current floor	{"elevator": "A", "floor": 2}
2023-08-12T17:10:44.772+0300	info	open doors	{"elevator": "B", "floor": 9}
2023-08-12T17:10:44.772+0300	info	open doors	{"elevator": "A", "floor": 2}
2023-08-12T17:10:46.783+0300	info	close doors	{"elevator": "A", "floor": 2}
2023-08-12T17:10:46.783+0300	info	close doors	{"elevator": "B", "floor": 9}
2023-08-12T17:10:50.240+0300	info	Request has been approved	{"elevator": "B", "fromFloor": 7, "toFloor": 0}
2023-08-12T17:10:50.240+0300	info	Request has been approved	{"elevator": "B", "fromFloor": 3, "toFloor": 0}
2023-08-12T17:10:50.240+0300	debug	current floor	{"elevator": "B", "floor": 9}
2023-08-12T17:10:50.750+0300	debug	current floor	{"elevator": "B", "floor": 8}
2023-08-12T17:10:51.266+0300	debug	current floor	{"elevator": "B", "floor": 7}
2023-08-12T17:10:51.768+0300	info	open doors	{"elevator": "B", "floor": 7}
2023-08-12T17:10:53.772+0300	info	close doors	{"elevator": "B", "floor": 7}
2023-08-12T17:10:53.772+0300	debug	current floor	{"elevator": "B", "floor": 6}
2023-08-12T17:10:54.286+0300	debug	current floor	{"elevator": "B", "floor": 5}
2023-08-12T17:10:54.802+0300	debug	current floor	{"elevator": "B", "floor": 4}
2023-08-12T17:10:55.304+0300	debug	current floor	{"elevator": "B", "floor": 3}
2023-08-12T17:10:55.806+0300	info	open doors	{"elevator": "B", "floor": 3}
2023-08-12T17:10:57.809+0300	info	close doors	{"elevator": "B", "floor": 3}
2023-08-12T17:10:57.809+0300	debug	current floor	{"elevator": "B", "floor": 2}
2023-08-12T17:10:58.310+0300	debug	current floor	{"elevator": "B", "floor": 1}
2023-08-12T17:10:58.816+0300	debug	current floor	{"elevator": "B", "floor": 0}
2023-08-12T17:10:59.320+0300	info	open doors	{"elevator": "B", "floor": 0}
2023-08-12T17:11:01.323+0300	info	close doors	{"elevator": "B", "floor": 0}
``````