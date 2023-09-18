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
* unit tests (done) https://plugins.jetbrains.com/plugin/17510-machinet-ai-gpt4--chatgpt


## Docker
```bash
docker build -t elevator . 
docker run --rm -p 6660:6660  --name elevator elevator:latest   
```
