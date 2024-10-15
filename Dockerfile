# Start from the official Golang base image
FROM golang:latest

# Set the working directory
WORKDIR /app

# Copy the Go mod and sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go app
RUN go build -o bin/gobank

# Expose the API port
EXPOSE 8080



# Run the executable
CMD ["./bin/gobank"]
