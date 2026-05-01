# Build the manager binary
# Use Alpine-based Go image for smaller size
FROM golang:1.23-alpine as builder

# Build arguments for versioning
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

WORKDIR /workspace

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

# Build with optimizations:
# -ldflags="-s -w" strips debug information
# CGO_ENABLED=0 creates static binary
# -trimpath removes file system paths from binary
# -X injects version information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA} -X main.BuildDate=${BUILD_DATE}" \
    -trimpath \
    -a -o manager main.go

# Use distroless as minimal base image
FROM gcr.io/distroless/static:nonroot
WORKDIR /

# Copy the binary from builder
COPY --from=builder /workspace/manager .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use non-root user
USER 65532:65532

ENTRYPOINT ["/manager"]
