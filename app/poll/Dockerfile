FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o poll_service ./app/poll/main.go

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/poll_service .

EXPOSE 8011

CMD ["./poll_service"]
