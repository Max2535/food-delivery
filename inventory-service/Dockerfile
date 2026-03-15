FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o api ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/worker/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/api .
COPY --from=builder /app/worker .

EXPOSE 3004

CMD ["./api"]
