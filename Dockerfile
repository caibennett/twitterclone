FROM golang:1.22.2-alpine

RUN apk add --no-cache build-base bash sqlite nodejs npm

WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN npm i tailwindcss -g
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest && go install github.com/a-h/templ/cmd/templ@latest
RUN tailwindcss --minify -i ./templ/input.css -o ./static/output.css && sqlc generate && templ generate
RUN go build -o bin .
RUN bash apply_fresh.sh

ENTRYPOINT "/app/bin"

EXPOSE 3000