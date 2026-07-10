FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY . .
RUN go mod tidy
RUN go build -o transactionhub ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/transactionhub .

EXPOSE 8080

CMD ["./transactionhub"]
