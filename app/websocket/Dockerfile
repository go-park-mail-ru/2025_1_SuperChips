FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build ./app/websocket/main.go

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app .

EXPOSE 8013

CMD ["./main"]
