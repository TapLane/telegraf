# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make bash

# Set working directory
WORKDIR /go/src/github.com/influxdata/telegraf

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code
COPY . .

# Remove problematic symlinks
RUN find . -type l -delete

# Build telegraf
RUN make build

# Runtime stage
FROM alpine:latest

# Install runtime dependencies including SNMP tools
RUN apk add --no-cache ca-certificates tzdata net-snmp-tools

# Copy the binary from builder
COPY --from=builder /go/src/github.com/influxdata/telegraf/telegraf /usr/bin/telegraf

# Create config directory
RUN mkdir -p /etc/telegraf /etc/telegraf/telegraf.d

# Expose default port (if using http_listener or similar)
EXPOSE 8125/udp 8092/udp 8094 8186

# Set telegraf as entrypoint
ENTRYPOINT ["/usr/bin/telegraf"]
CMD ["--config", "/etc/telegraf/telegraf.conf", "--config-directory", "/etc/telegraf/telegraf.d"]
