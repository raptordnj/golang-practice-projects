FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o product-service main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/product-service .
EXPOSE 8081
CMD ["./product-service"]
