# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY ./common-service/go.mod ./common-service/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY ./common-service/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o common ./common.go

# Run stage
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary and config from builder
COPY --from=builder /app/common .
COPY --from=builder /app/etc/common.yaml ./etc/

USER appuser

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["./common"]
