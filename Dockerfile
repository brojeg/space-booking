# Use the official Golang image as the base image
FROM golang:1.23-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/api/main.go

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the Go app
CMD ["./cmd/api/main.go"]