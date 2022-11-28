BINARY=api
VERSION=0.1.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

run:
	go run main.go

build:
	go mod tidy
	go build ./...

build-docker:
	go mod tidy
	go build ${LDFLAGS} -o ${BINARY} main.go

build-image:
	docker build --tag gracefulshutdown .

run-image:
	docker run -p 8081:8000 gracefulshutdown

delete-image:
	docker image rm gracefulshutdown:latest