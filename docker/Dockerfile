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

# Copy Swagger UI static files (assumes `swagger-ui` is in your project root)
COPY swagger-ui /app/swagger-ui

# Stage 2: Final image
FROM alpine:3.18

# Install redis-cli and essential tools
RUN apk add --no-cache \
    bash \
    vim \
    curl \
    wget \
    less \
    redis

# Set up a non-root user and group
RUN addgroup -S keyholder && adduser -S admin -G keyholder

# Set working directory and copy built files from builder stage
WORKDIR /app
COPY --from=builder /app/keyhouse /app/keyhouse
COPY --from=builder /app/swagger-ui /app/swagger-ui
COPY --from=builder /app/pkg/pb/docs /app/docs

# Grant ownership and restrict permissions for security
RUN chown -R admin:keyholder /app && chmod -R 755 /app

# Switch to non-root user
USER admin

# Run the binary
CMD ["/app/keyhouse"]
