.PHONY: build run test lint fmt clean check

build:
	go build -o bin/tuime ./...

run:
	go run main.go

test:
	go test -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -rf bin/

check: fmt lint test build
	@echo "All checks passed!"
