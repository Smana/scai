.PHONY: build install test clean lint run-example

BINARY_NAME=scia
GO=go

build:
	$(GO) build -o $(BINARY_NAME) .

install: build
	$(GO) install

test:
	$(GO) test -v ./...

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

lint:
	golangci-lint run

run-example:
	./$(BINARY_NAME) deploy \
		"Deploy this Flask app on AWS" \
		https://github.com/Arvo-AI/hello_world

.DEFAULT_GOAL := build
