# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY api/ api/
COPY cmd/ cmd/
COPY internal/ internal/

# Build the controller
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager ./cmd/manager/main.go

# Final stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/manager .

# Copy dependency files for vulnerability scanning
COPY --from=builder /workspace/go.mod /workspace/go.sum ./

# Use non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]
