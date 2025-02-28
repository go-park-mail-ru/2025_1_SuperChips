FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o main ./cmd/main.go

FROM alpine:latest  

COPY --from=builder /app/main /main

ENV PORT=8080
EXPOSE $PORT

ENTRYPOINT ["/main"]
