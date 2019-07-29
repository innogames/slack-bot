
.PHONY: clean docker-build test test-coverage test-bench mocks

all: clean dep test build/slack-bot

build/slack-bot:
	mkdir -p build/
	go build -o build/slack-bot cmd/bot/main.go

clean:
	rm -rf build/

dep:
	go mod vendor
	go get github.com/vektra/mockery/

docker-build:
	docker build . --force-rm -t slack-bot:latest

test:
	go test ./...

test-bench:
	go test -bench . ./... -benchmem

test-coverage:
	mkdir -p build && go test ./... -coverpkg=./... -cover -coverprofile=./build/cover.out -covermode=atomic && go tool cover -html=./build/cover.out -o ./build/cover.html

mocks: mocks/SlackClient.go mocks/Stash.go mocks/JenkinsJob.go mocks/JenkinsClient.go

mocks/SlackClient.go:
	$$GOPATH/bin/mockery -dir client/ -name SlackClient

mocks/Stash.go:
	$$GOPATH/bin/mockery -dir vendor/github.com/xoom/stash -name Stash

mocks/JenkinsJob.go:
	$$GOPATH/bin/mockery -dir client/jenkins -name Job

mocks/JenkinsClient.go:
	$$GOPATH/bin/mockery -dir client/jenkins -name Client
