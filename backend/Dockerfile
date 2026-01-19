# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build deps if needed
RUN apk add --no-cache git

# Copy Mod files
COPY go.mod ./
RUN go mod download

# Copy Source
COPY . .

# Build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o registry-backend main.go

# Run Stage
FROM alpine:3.19

WORKDIR /app

# Install Runtime Deps & Trivy
RUN apk add --no-cache ca-certificates curl \
    && curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin

COPY --from=builder /app/registry-backend .
COPY --from=builder /app/migrations ./migrations

# Expose Port
EXPOSE 5000

# Run
CMD ["./registry-backend"]
