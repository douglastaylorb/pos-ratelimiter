FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy .env file if it exists
COPY .env* ./

EXPOSE 8080

CMD ["./main"]