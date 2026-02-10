FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -o /user-service ./cmd/user-service

FROM alpine:3.19
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /user-service .
EXPOSE 8081
CMD ["./user-service"]
