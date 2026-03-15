FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 3003

CMD ["./main"]
