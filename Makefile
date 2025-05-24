BINARY=osv-crawler
all: build

build:
	go build -o $(BINARY) ./cmd/osv-crawler

run:
	./$(BINARY) --env-file .env

test:
	go test ./...
