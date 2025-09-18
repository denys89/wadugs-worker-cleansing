# syntax=docker/dockerfile:1.3
# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Use the secret to access the SSH key
RUN --mount=type=secret,id=ssh_key \
    mkdir -p /root/.ssh && \
    cat /run/secrets/ssh_key > /root/.ssh/id_ed25519 && \
    chmod 600 /root/.ssh/id_ed25519

# Install Git
RUN apk update && apk add --no-cache git coreutils openssh-client

RUN git config --global url."git@gitlab.com:".insteadOf https://gitlab.com/ \
    && git config --global url."git@github.com:".insteadOf https://github.com/ \
    && ssh-keyscan gitlab.com >> ~/.ssh/known_hosts \
    && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go env -w GOPRIVATE=github.com/harryosmar && go mod download

# Copy the source code to the container
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Stage 2: Create the final image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose the port (if needed)
# EXPOSE 8080

# Command to run the binary
CMD ["./main"]