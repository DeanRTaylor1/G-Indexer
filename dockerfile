# First stage: Build the Go application
FROM golang:1.20-bullseye AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY . .
RUN go build -o bin/main .

# Second stage: Run the Go application
# Second stage: Create the final image
FROM debian:bullseye-slim

# Set the working directory
WORKDIR /app

# Copy the binary from the first stage
COPY --from=builder /app/bin/main .


# Expose the server's port
EXPOSE 8080

# Run the server
CMD ["./main"]
