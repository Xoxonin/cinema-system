# Build stage
FROM golang:1.26.1-alpine3.23 AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build the application
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:3.23

# Fix CVE-2026-22184 and CVE-2026-27171
RUN apk update && apk add --no-cache --upgrade zlib

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8084
CMD ["./main"]
