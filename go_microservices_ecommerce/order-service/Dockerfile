FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o order-service main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8082
CMD ["./order-service"]
