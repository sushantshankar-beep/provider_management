# ---------- Build stage ----------
FROM golang:1.23-alpine as builder

WORKDIR /app

# Install git (needed for go mod)
RUN apk add --no-cache git

# Copy go mod files first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o provider_management ./cmd/app

# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

# Copy binary only
COPY --from=builder /app/provider_management .

# Expose app port
EXPOSE 6600

# Run app
CMD ["./provider_management"]
