# Multi-stage Dockerfile for the backend
# Builder stage: compile a static Go binary
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates build-base

WORKDIR /src

# Cache go modules (use repository root go.mod/go.sum)
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://proxy.golang.org,direct && go mod download

# Copy backend source into the build context and build the server from backend/cmd/server
COPY backend ./backend

# Build statically
ENV CGO_ENABLED=0
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/server ./backend/cmd/server

# Runtime stage: small alpine image
FROM alpine:3.18

# Install runtime dependencies (certs for TLS if needed)
RUN apk add --no-cache ca-certificates

# Create non-root user for security
RUN addgroup -S app && adduser -S app -G app

# Copy the built binary from the builder stage
COPY --from=builder /app/server /app/server

# Set working directory, ownership and expose port
WORKDIR /app
RUN chown app:app /app/server
USER app

EXPOSE 8080

# Use environment variables for configuration (read by the app)
ENV PORT=8080
ENV GIN_MODE=release

# Start the server
CMD ["/app/server"]
