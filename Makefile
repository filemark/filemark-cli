.PHONY: build-cli test

build-cli:
	@echo "Building service"
	go build -o filemark-cli

test:
	@echo "Running unit tests"
	@go test -cover `go list ./... | grep -v int_test`
