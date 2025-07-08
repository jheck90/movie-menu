# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o movie-menu .

# Runtime stage
FROM alpine:latest

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/movie-menu .
COPY --from=builder /app/web ./web

# Optional: copy default config if needed
# COPY config ./config

# Allow mounting of external config or .env files
VOLUME ["/app/config"]

EXPOSE 8080

ENTRYPOINT ["./movie-menu"]
