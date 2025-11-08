# Build stage
FROM golang:1.25.1-alpine AS builder

# Add Maintainer Info
LABEL maintainer="Mohammad Kaium <mohammadkaiom79@gmail.com>"

# Install git and ca-certificates (needed for swagger and HTTPS)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install swag CLI for generating swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
RUN swag init

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ride_engine .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/ride_engine .
COPY --from=builder /app/docs ./docs

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./ride_engine", "serve"]