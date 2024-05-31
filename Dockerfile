FROM golang:1.22.2-alpine

RUN apk add --no-cache build-base bash sqlite

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o bin .
RUN bash apply_fresh.sh

ENTRYPOINT "/app/bin"

EXPOSE 3000