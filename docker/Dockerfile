# Stage 1: Build the binary
FROM golang:1.22.5-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Cache dependencies for faster rebuilds
COPY go.mod go.sum ./
RUN go mod tidy && \
    go mod vendor

# Copy source code and build the app
COPY . .
RUN go build -o keyhouse ./cmd/main.go

# Stage 2: Final image
FROM alpine:3.18

# Set up a non-root user and group
RUN addgroup -S keyholder && adduser -S admin -G keyholder

# Set working directory and copy built files from builder stage
WORKDIR /app
COPY --from=builder /app/keyhouse /app/keyhouse
COPY --from=builder /app/pkg/pb/docs /app/docs

# Grant ownership and restrict permissions for security
RUN chown -R admin:keyholder /app && chmod -R 755 /app

# Switch to non-root user
USER admin

# Expose necessary ports
EXPOSE 8080 30100
EXPOSE 8081 30101

# Run the binary
CMD ["/app/keyhouse"]
