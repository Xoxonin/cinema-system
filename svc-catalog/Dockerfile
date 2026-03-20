FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/catalog ./cmd/catalog

FROM alpine:3.21
RUN addgroup -S app && adduser -S app -G app
WORKDIR /app

COPY --from=builder /out/catalog ./catalog
USER app

EXPOSE 8081
CMD ["./catalog"]
