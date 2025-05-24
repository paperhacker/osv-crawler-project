FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git && go mod tidy && go build -o osv-crawler ./cmd/osv-crawler

FROM alpine:3.18
WORKDIR /crawler
COPY --from=builder /app/osv-crawler .
ENTRYPOINT ["./osv-crawler"]
