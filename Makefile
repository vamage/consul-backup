.PHONY: all deps format

all: deps format
	@go build

deps:
	@echo "--> Installing build dependencies"
	@go get -d -v ./...

format: deps
	@echo "--> Running go fmt"
	@go fmt
