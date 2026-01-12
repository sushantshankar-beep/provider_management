# ---------- Build stage ----------
FROM golang:1.24-alpine AS builder


WORKDIR /app

# Install git (for go mod)
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o provider_management ./cmd/app

# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

# CA certs for MongoDB Atlas (IMPORTANT)
RUN apk add --no-cache ca-certificates

# Copy binary only
COPY --from=builder /app/provider_management .

# Expose container port (documentation only)
EXPOSE 6600

# ✅ IMPORTANT: exec form (inherits env correctly)
CMD ["./provider_management"]
