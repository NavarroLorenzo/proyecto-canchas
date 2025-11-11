# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar git
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download -x

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env* ./

# Expose port
EXPOSE 8082

# Run the application
CMD ["./main"]