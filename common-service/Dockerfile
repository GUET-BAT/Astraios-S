# Build stage
FROM golang:1.25-alpine AS builder

# Set Go proxy for China network
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Copy go mod files (go.sum may be incomplete)
COPY ./common-service/go.mod ./common-service/go.sum ./

# Copy source code (needed for go mod tidy to resolve imports)
COPY ./common-service/ ./

# Update go.sum and download dependencies
RUN go mod tidy

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
