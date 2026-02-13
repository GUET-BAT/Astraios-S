# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Workspace and module sources
COPY go.work go.work.sum ./
COPY gateway-service/ gateway-service/
COPY common-service/ common-service/
COPY user-service/ user-service/
COPY global/golang/ global/golang/

# Sync workspace dependencies
RUN go work sync

WORKDIR /app/user-service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user ./user.go

# Run stage
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/user-service/user .
COPY --from=builder /app/user-service/etc/user.yaml ./etc/

USER appuser

EXPOSE 8080

ENTRYPOINT ["./user"]
