.PHONY: all build build-web build-go test clean run

BINARY_NAME=bin/skupilot

all: build

build-web:
	cd web && npm install && npm run build

build-go:
	go build -o $(BINARY_NAME) ./cmd/skupilot

build: build-web build-go

test:
	go test -v -race ./...

clean:
	rm -rf bin/ web/dist/ catalog.db

run: build
	./$(BINARY_NAME) serve --port 8080 --db ./catalog.db
