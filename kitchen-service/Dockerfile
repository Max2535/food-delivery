FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build both api and worker
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/worker .

EXPOSE 3001

CMD ["./api"]
