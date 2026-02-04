# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY gateway-service/go.mod gateway-service/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY gateway-service/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./api/gateway.go

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/gateway .
COPY --from=builder /app/api/etc/gateway.yaml ./etc/

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./gateway"]
