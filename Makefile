
.PHONY: clean docker-build test test-coverage test-bench mocks run

all: clean dep test build/slack-bot

build/slack-bot:
	mkdir -p build/
	go build -o build/slack-bot cmd/bot/main.go

run:
	go run cmd/bot/main.go

clean:
	rm -rf build/

dep:
	go mod vendor

docker-build:
	docker build . --force-rm -t brainexe/slack-bot:latest

test:
	go test ./... -race

test-bench:
	go test -bench . ./... -benchmem

test-coverage:
	mkdir -p build && go test ./... -coverpkg=./... -cover -coverprofile=./build/cover.out -covermode=atomic && go tool cover -html=./build/cover.out -o ./build/cover.html

mocks: dep
	go get github.com/vektra/mockery/
	go generate ./...
