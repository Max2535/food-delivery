# Stage 1: Build Go plugin using KrakenD's official plugin builder
# (must use same Go version as KrakenD runtime)
FROM devopsfaith/krakend-plugin-builder:latest AS builder

WORKDIR /app
COPY gateway/plugin/ .

RUN go build -buildmode=plugin -o /correlation-id.so .

# Stage 2: KrakenD runtime with the compiled plugin
FROM devopsfaith/krakend:latest

# Copy the compiled plugin into KrakenD's plugin directory
COPY --from=builder /correlation-id.so /etc/krakend/plugins/correlation-id.so
