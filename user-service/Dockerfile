# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Workspace and module sources
COPY go.work ./
COPY common-service/ common-service/
COPY user-service/ user-service/

WORKDIR /app/user-service

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user ./user.go

# Run stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/user-service/user .
COPY --from=builder /app/user-service/etc/user.yaml ./etc/

EXPOSE 8080

ENTRYPOINT ["./user"]
