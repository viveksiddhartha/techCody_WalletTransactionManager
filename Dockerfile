# Use a golang base image with Go installed
FROM golang:1.20.2-alpine3.17 AS build


# Set the working directory
WORKDIR /app

# Copy the Go module files and download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /app/main .

# Use a minimal base image for running the application
FROM alpine:3.17

# Set the working directory
WORKDIR /app

# Copy the application binary from the build image
COPY --from=build /app/main .

# Expose the default HTTP port
EXPOSE 8080

# Start the application
CMD ["./walletManager"]
