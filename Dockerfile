# STAGE 1: Build
FROM golang:1.25-alpine AS builder

# Install git (required for fetching dependencies)
RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies (Docker Layer Caching)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# We build the simulator for now. Later we will change this to the server.
RUN CGO_ENABLED=0 GOOS=linux go build -o tardigo-sim ./cmd/simulator

# STAGE 2: Run
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/tardigo-sim .

# Command to run
CMD ["./tardigo-sim"]