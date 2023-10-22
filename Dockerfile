FROM golang:1.21

WORKDIR /app

COPY . .

RUN go fmt ./...

RUN go vet ./...

RUN go build -o main ./cmd/api

COPY --from=builder /app/main .

CMD ["./main"]
