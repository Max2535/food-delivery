FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/main .
# Optionally copy .env if used, or pass via docker-compose environment
# COPY --from=builder /app/.env .

EXPOSE 3000

CMD ["./main"]
