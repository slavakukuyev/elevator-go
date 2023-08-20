# Stage 1: Build the Go application
FROM golang:1.21-alpine AS base

# Set the working directory
WORKDIR /src

# Set environment variables for Go module downloads
ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOPROXY=proxy.golang.org,direct

COPY go.mod go.sum /src/
RUN go mod download

COPY *.go .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o elevator

# Stage 2: Create the final image using a smaller base image
FROM alpine:3.18

# Set the working directory
WORKDIR /src

# Create the /go/bin directory
RUN mkdir -p /go/bin

# Copy the compiled binary from the build stage
COPY --from=base /src/elevator /go/bin/


EXPOSE 8080
# Run the Go application
CMD ["/go/bin/elevator"]
