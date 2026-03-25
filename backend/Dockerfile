FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git (needed for private modules if any)
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-s -w' -o /app/server ./cmd/api

# Runtime image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
