GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

# Help about any target
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Available Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: init
# Initialize the required tools & dependencies
init:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

	kratos upgrade
	buf dep update

.PHONY: build
# Build the project. The build output will be in the `build` directory
build:
	go mod tidy
	rm -rf ./build
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o ./build/ ./...

.PHONY: build-image
# Build the docker image
build-image:
	docker build -t ${MODULE_PREFIX}:${VERSION} -f Dockerfile .
	docker tag $(MODULE_PREFIX):$(VERSION) $(MODULE_PREFIX):latest

.PHONY: generate
# Generate the proto files & wire dependencies
generate:
	buf generate
	wire gen ./...
