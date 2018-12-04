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
	go test -v -cover ./...

setup:
	go get github.com/mongodb/mongo-go-driver/mongo
	docker pull mongo:4.0.4
	docker build -t rech.haeser.daniel-data-integration-challenge .
