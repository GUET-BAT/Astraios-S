# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Workspace and module sources
COPY go.work go.work.sum ./
COPY user-service/ user-service/
COPY gateway-service/ gateway-service/

WORKDIR /app/gateway-service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gateway ./gateway.go

# Run stage
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/gateway-service/gateway .
COPY --from=builder /app/gateway-service/etc/gateway.yaml ./etc/

USER appuser

EXPOSE 8888

ENTRYPOINT ["./gateway"]
