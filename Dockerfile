FROM golang:1.16-alpine as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download
COPY . /app
RUN go build .

EXPOSE 8080
ENTRYPOINT ["./chirpbird"]