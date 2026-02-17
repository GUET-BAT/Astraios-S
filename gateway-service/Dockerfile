# Build stage
FROM golang:1.25-alpine AS builder

# Set Go proxy for China network
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# Workspace and module sources
COPY go.work go.work.sum ./
COPY common-service/go.mod common-service/go.sum ./common-service/
COPY user-service/go.mod user-service/go.sum ./user-service/
COPY gateway-service/go.mod gateway-service/go.sum ./gateway-service/
COPY global/golang/go.mod ./global/golang/

# Sync workspace dependencies
RUN go work sync

# Copy source code
COPY common-service/ common-service/
COPY user-service/ user-service/
COPY gateway-service/ gateway-service/
COPY global/golang/ global/golang/

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
