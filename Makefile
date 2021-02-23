
.PHONY: clean docker-build test test-coverage test-bench mocks run dep lint

all: clean dep test build/slack-bot

FLAGS = -trimpath -ldflags="-s -w -X github.com/innogames/slack-bot/bot.Version=$(shell git describe --tags)"

build/slack-bot: dep
	mkdir -p build/
	go build $(FLAGS) -o build/slack-bot cmd/bot/main.go

run: dep
	go run $(FLAGS) cmd/bot/main.go

run-cli: dep
	go run cmd/cli/main.go

clean:
	rm -rf build/

dep:
	go mod vendor

lint:
	golangci-lint run

docker-build:
	docker build . --force-rm -t brainexe/slack-bot:latest

test: dep
	go test ./... -race

test-bench:
	go test -bench . ./... -benchmem

test-coverage: dep
	mkdir -p build
	go test ./... -coverpkg=./... -cover -coverprofile=./build/cover.out -covermode=atomic
	go tool cover -html=./build/cover.out -o ./build/cover.html
	echo see ./build/cover.html

mocks: dep
	go get github.com/vektra/mockery/v2/.../
	go generate ./...
