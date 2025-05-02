FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v ./app/main/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app .

COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/static/ ./static/
COPY --from=builder /app/docs ./docs/

EXPOSE 8080

ENTRYPOINT ["./main"]
