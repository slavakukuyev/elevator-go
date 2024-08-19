# Stage 1: Build the Go binary
FROM golang:1.21-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/out/elevator-go ./cmd/server

# Stage 2: Run the Go binary
FROM alpine:3.19

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/out/elevator-go .

# Expose port if needed (for example, if your app listens on port 6660)
EXPOSE 6660

# Command to run the executable
CMD ["./elevator-go"]
