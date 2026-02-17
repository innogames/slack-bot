
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
	go run -tags pprof $(FLAGS) cmd/bot/*.go

run-cli: dep
	@test -f config.yaml || (echo "please create a config.yaml first. Hint: check the config.example.yaml" && exit 1)
	go run $(FLAGS) cmd/cli/main.go -config config.yaml

run-cli-config:
	go run cmd/cli/main.go -config config.yaml

clean:
	rm -rf build/

# download go dependencies into ./vendor/
dep:
	@go mod vendor

lint:
	go fix ./...
	golangci-lint run --fix

docker-build:
	docker build . --force-rm -t brainexe/slack-bot:latest

docker-push:
	docker push brainexe/slack-bot:latest

test: dep
	go test ./...

test-race: dep
	go test ./... -race

test-bench:
	go test -bench . ./... -benchmem

test-coverage: dep
	@mkdir -p build
	go test ./... -coverpkg=./... -cover -coverprofile=./build/cover.out -covermode=atomic
	go tool cover -html=./build/cover.out -o ./build/cover.html
	@go tool cover -func ./build/cover.out | grep total | awk '{print "Total Coverage: " $$3 " see ./build/cover.html"}'

# build mocks for testable interfaces into ./mocks/ directory
mocks: dep
	command -v mockery || go install github.com/vektra/mockery/v2@latest
	go generate ./...

# live reload, see https://github.com/cosmtrek/air
run-live-reload:
	command -v air || go install github.com/cosmtrek/air@latest
	air
