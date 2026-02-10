FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -o /order-service ./cmd/order-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /order-service .
EXPOSE 8091
CMD ["./order-service"]
