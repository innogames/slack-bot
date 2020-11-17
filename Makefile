
.PHONY: clean docker-build test test-coverage test-bench mocks run dep

all: clean dep test build/slack-bot

build/slack-bot:
	mkdir -p build/
	go build -o build/slack-bot -ldflags="-s -w" cmd/bot/main.go

run:
	go run cmd/bot/main.go

clean:
	rm -rf build/

dep:
	go mod vendor

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
