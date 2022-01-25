
.PHONY: clean docker-build test test-coverage test-bench mocks run run-cli dep lint air

all: test build/slack-bot build/cli

FLAGS = -trimpath -ldflags="-s -w -X github.com/innogames/slack-bot/v2/bot/version.Version=$(shell git describe --tags)"

build/slack-bot: dep
	@mkdir -p build/
	go build $(FLAGS) -o build/slack-bot cmd/bot/main.go

build/cli: dep
	@mkdir -p build/
	go build $(FLAGS) -o build/cli cmd/cli/main.go

run: dep
	go run $(FLAGS) cmd/bot/main.go

run-cli:
	go run $(FLAGS) cmd/cli/main.go

run-cli-config:
	go run cmd/cli/main.go -config config.yaml

clean:
	rm -rf build/

# download go dependencies into ./vendor/
dep:
	@go mod vendor

lint:
	golangci-lint run

docker-build:
	docker build . --force-rm -t brainexe/slack-bot:latest

test: dep
	go test ./... -race

test-bench:
	go test -bench . ./... -benchmem

test-coverage: dep
	@mkdir -p build
	go test ./... -coverpkg=./... -cover -coverprofile=./build/cover.out -covermode=atomic
	go tool cover -html=./build/cover.out -o ./build/cover.html
	@echo see ./build/cover.html

# build mocks for testable interfaces into ./mocks/ directory
mocks: dep
	go get github.com/vektra/mockery/v2/.../
	go generate ./...

# live reload
air:
	command -v air || curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	air
