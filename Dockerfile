# Pull golang alpine to build binary
FROM golang:alpine as builder

# Update the package list and install the required packages
RUN apk add --no-cache gcc musl-dev make bash

# Set the working directory
WORKDIR /app

# Copy the source code and Makefile into the container
COPY . .

# Build the application
RUN make build-app

# RUN CGO_ENABLED=1 go build -a -o bin/magic-chain -ldflags="-w -s" ./cmd/magic-chain/

# Use a lightweight base image for the final stage
FROM alpine:latest
COPY --from=builder /app/bin/magic-chain .

# Copy the binary from the builder stage
# COPY --from=builder /app/bin/magic-chain /bin/magic-chain

# Run app and expose api and metrics ports

# API
EXPOSE 4001 4001

# Metrics
# EXPOSE 7300

# Run app
CMD ["./magic-chain"]

