# STAGE 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build Simulator
RUN CGO_ENABLED=0 GOOS=linux go build -o tardigo-sim ./cmd/simulator

# Build API Server (NEW)
RUN CGO_ENABLED=0 GOOS=linux go build -o tardigo-api ./cmd/api

# STAGE 2: Run
FROM alpine:latest
WORKDIR /root/

# Copy both binaries
COPY --from=builder /app/tardigo-sim .
COPY --from=builder /app/tardigo-api .

# We don't set a default CMD anymore because docker-compose will decide which one to run