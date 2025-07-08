FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o movie-menu

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /app/movie-menu /movie-menu
COPY web /web
EXPOSE 8080
ENTRYPOINT ["/movie-menu"]
