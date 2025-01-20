# Step 1: Use the official Go image as a build stage
FROM golang:latest AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Go modules manifests
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod tidy

# Copy the source code into the container
COPY . .

# Step 2: Build the Go application
RUN go build -o myapp .

# Step 3: Create a smaller image for production
FROM debian:bullseye-slim

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the pre-built binary from the builder stage
COPY --from=builder /app/myapp .

# Expose the port the app runs on
EXPOSE 8080

# Run the Go app
CMD ["./myapp"]
