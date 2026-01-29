FROM golang:alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application with verbose output
RUN go build -v -o expenseowl ./cmd/expenseowl

# Use a minimal alpine image for running
FROM alpine:latest

WORKDIR /app

# Create data directory if not exists
RUN mkdir -p /app/data

# Copy the binary from builder
COPY --from=builder /app/expenseowl .

# Expose the default port
EXPOSE 8080

# Run the server
CMD ["./expenseowl"]
