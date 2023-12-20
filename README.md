# elevator-go
Elevator Control System
This repository contains an Elevator Control System written in Go that simulates the operation of multiple elevators within a building.


## TODO
* create env (done)
* add GitHub Actions (done)
* add Logic of floor under 0 like parking (done)
* use zap logger (done)
* create docker image (done)
* create manager algorithm on base of all elevators (done)
* add handler to creating new elevators for the system (done)
* add http server for debug different cases (done)
* front page (not sure  with which framework)
* investigate delay on start (why?) (done)
* unit tests (done) 
* support client requests in gRPC
* support Prototype design pattern to clone elevators
* create default endpoint to provide an available endpoints to api (like GET api.elevator.com)
* Create default files structure by golang conventions
* OAuth / OAuth 2.0
* Versioning of API in URL (example: https://elevator.com/v1/newLift/...)

## Docker
```bash
docker build -t elevator . 
docker run --rm -p 6660:6660  --name elevator elevator:latest 
```
