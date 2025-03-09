# Use the official golang image as the base image
FROM golang:1.23-bookworm AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the auth service source code
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install necessary dependencies for the application to run
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /auth-service .

# Expose the port the auth service runs on
EXPOSE 50051

# Command to run the application
CMD ["./auth-service"]