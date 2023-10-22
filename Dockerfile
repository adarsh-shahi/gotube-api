FROM golang:1.21

WORKDIR /app

COPY . /app

RUN go fmt ./...

RUN go vet ./...

RUN go build -o ./cmd/api

CMD ["./api"]
