FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN [ -f go.sum ] || touch go.sum

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest  

COPY --from=builder /app/main /main

ENV PORT=8080
EXPOSE $PORT

ENTRYPOINT ["/main"]
