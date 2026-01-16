# Start from the official Golang image for building
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o by_who

# Use a minimal image for running
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /app/by_who .

# Expose port (change 8080 to your app's port if different)
EXPOSE 8082

# Command to run the binary
CMD ["./by_who"]
