# elevator-go
Elevator Control System
This repository contains an Elevator Control System written in Go that simulates the operation of multiple elevators within a building.



* create env (done)
* add GitHub Actions
* add Logic of floor under 0 like parking (done)
* use zap logger (done)
* create docker image (done)
* create manager algorithm on base of all elevators (done)
* add handler to creating new elevators for the system
* add http server for debug different cases (done)
* front page (not sure  with which framework)
* investigate delay on start (why?) (done)
* tests


## Docker
```bash
docker build -t elevator . 
docker run --rm -p 1010:1010  --name elevator elevator:latest   
```
