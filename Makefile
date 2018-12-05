#Challenge Makefile

start:
	docker network create rech.haeser.daniel-network
	docker run -d -p 27017:27017 --net rech.haeser.daniel-network --rm --name rech.haeser.daniel-mongo mongo:4.0.4
	docker run -d -p 8080:8080   --net rech.haeser.daniel-network --rm --name rech.haeser.daniel-data-integration-challenge rech.haeser.daniel-data-integration-challenge

stop:
	- docker stop rech.haeser.daniel-data-integration-challenge
	- docker stop rech.haeser.daniel-mongo
	- docker network rm rech.haeser.daniel-network

check:
	docker run -d -p 27018:27017 --rm --name rech.haeser.daniel-mongo_test mongo:4.0.4
	- go test -v -cover ./...
	docker stop rech.haeser.daniel-mongo_test

setup:
	go get github.com/mongodb/mongo-go-driver/mongo
	go get github.com/go-chi/chi
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	docker pull mongo:4.0.4
	docker build -t rech.haeser.daniel-data-integration-challenge .

spec:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	${GOPATH}/bin/swagger generate spec -o swagger.json
	${GOPATH}/bin/swagger serve -F swagger swagger.json
