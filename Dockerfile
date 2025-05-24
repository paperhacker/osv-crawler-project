# ---- Stage 1: Build Go 1.23.9 from source ----
FROM alpine:3.18 AS golang-builder

ARG GO_VERSION=1.23.9

# Install required tools, including bash for make.bash
RUN apk add --no-cache curl tar gcc musl-dev go git bash

WORKDIR /usr/local

# Download and extract Go source
RUN curl -LO https://go.dev/dl/go${GO_VERSION}.src.tar.gz && \
    tar -xzf go${GO_VERSION}.src.tar.gz && \
    rm go${GO_VERSION}.src.tar.gz

# Build Go from source
WORKDIR /usr/local/go/src
RUN ./make.bash

# ---- Stage 2: Build your Go app ----
FROM alpine:3.18 AS builder

# Copy Go 1.23.9 toolchain
COPY --from=golang-builder /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:$PATH"

# Set up working directory
WORKDIR /app

# Install git (required by go mod)
RUN apk add --no-cache git

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy and build the app
COPY . .
RUN go build -o osv-crawler ./cmd/osv-crawler

# ---- Stage 3: Minimal runtime image ----
FROM alpine:3.18

WORKDIR /crawler
COPY --from=builder /app/osv-crawler .

ENTRYPOINT ["./osv-crawler"]

