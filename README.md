# elevator-go
Golang elevator real system implementation



* create env for elevators values
* add GitHub Actions
* add Logic of floor under 0 - like parking
* use zap logger (done)
* create docker image (done)
* create manager algorithm on base of all elevators (done)
* support min, max for each elevator separately. manager has to handle it
* add http server for debug different cases (done)
* front page (not sure  with which framework)
* investigate delay on start (why?)


## Docker
```bash
docker build -t elevator . 
docker run --rm -p 1010:1010  --name elevator elevator:latest   
```
