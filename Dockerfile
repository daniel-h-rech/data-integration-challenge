FROM golang:1.11.2

RUN go get github.com/mongodb/mongo-go-driver/mongo

ENV PROJECT_NAME data-integration-challenge
ENV PROJECT_ROOT ${GOPATH}/src/github.com/daniel-h-rech/${PROJECT_NAME}

RUN mkdir -p ${PROJECT_ROOT}

WORKDIR ${PROJECT_ROOT}

COPY . ${PROJECT_ROOT}

RUN go build

ENTRYPOINT ./${PROJECT_NAME} q1_catalog.csv rech.haeser.daniel-mongo:27017
