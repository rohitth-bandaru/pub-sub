# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pub-sub .

# Copy environment file
COPY env .env

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/pub-sub .

# Expose port
EXPOSE 8080

# Set environment variables
ENV PORT=8080
ENV HOST=0.0.0.0
ENV MAX_MESSAGES_PER_TOPIC=1000
ENV MAX_PUBLISH_RATE=100

# Run the application
CMD ["./pub-sub"]
