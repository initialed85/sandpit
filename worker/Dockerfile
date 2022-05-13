FROM golang:1.16.8-stretch

COPY go.mod /srv/go.mod
COPY go.sum /srv/go.sum

WORKDIR /srv/

RUN go mod download

COPY . /srv/

CMD go run cmd/main.go
