# Start from the official Golang image
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o openpm main.go

# Final lightweight image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/openpm ./openpm
CMD ["/app/openpm"]
