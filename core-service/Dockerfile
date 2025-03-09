# Use the official golang image as the base image
FROM golang:1.23-bookworm AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the core service source code
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /core-service ./cmd/main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install necessary dependencies for the application to run
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /core-service .

# Copy the registry.json file to the root of the container
COPY --from=builder /app/internal/apiserver/registry.json /registry.json

# Create a symbolic link to make the relative path work
RUN mkdir -p /app/../internal/apiserver
RUN ln -s /registry.json /app/../internal/apiserver/registry.json

# Expose the port the core service runs on
EXPOSE 8080

# Command to run the application
CMD ["./core-service"]