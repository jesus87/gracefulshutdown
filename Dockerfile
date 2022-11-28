# syntax=docker/dockerfile:1

#Build
FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /gracefulshutdown

EXPOSE 8000 8000

CMD [ "/gracefulshutdown" ]