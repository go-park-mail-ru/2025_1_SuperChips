FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o auth_service ./auth_service/app/main.go

RUN ls -l /app

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/auth_service .

EXPOSE 8010

CMD ["./main"]
