FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/movie-service ./cmd/movie-service

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/movie-service .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/dbconfig.yml .

EXPOSE 8080
EXPOSE 8081

CMD ["./movie-service"]
