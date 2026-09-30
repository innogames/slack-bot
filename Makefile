
.PHONY: clean docker-build test test-coverage test-bench mocks run run-cli dep lint air build-custom

all: test build/slack-bot build/cli

FLAGS = -trimpath -ldflags="-s -w -X github.com/innogames/slack-bot/v2/bot/version.Version=$(shell git describe --tags)"

# config which is used by "make run", "make run-cli" and "make build-custom", e.g. "make run CONFIG=config/"
CONFIG ?= config.yaml
GOEXE := $(shell go env GOEXE)

# the official plugins in ./plugins/ are own Go modules, see docs/plugins.md
PLUGIN_DIRS = $(patsubst %/go.mod,%,$(wildcard plugins/*/go.mod))

# executes the given command in each plugin module, e.g. $(call in-plugins,go test ./...)
define in-plugins
$(foreach dir,$(PLUGIN_DIRS),cd $(dir) && $(1)
)
endef

build/slack-bot: dep
	@mkdir -p build/
	go build $(FLAGS) -o build/slack-bot cmd/bot/main.go

build/cli: dep
	@mkdir -p build/
	go build $(FLAGS) -o build/cli cmd/cli/main.go

build/slack-bot-builder: dep
	@mkdir -p build/
	go build $(FLAGS) -o build/slack-bot-builder cmd/slack-bot-builder/main.go

# bot and cli binary including all plugins of ./plugins/
build/slack-bot-full: dep
	go run ./cmd/slack-bot-builder -config plugins/all.yaml -core . -output build/slack-bot-full -cli-output build/cli-full

# bot and cli binary including the plugins of the config, based on the local slack-bot
build-custom: dep
	go run ./cmd/slack-bot-builder -config $(CONFIG) -core . -output build/slack-bot-custom -cli-output build/cli-custom

# builds and starts the bot including the plugins of the config, with the pprof server
run: dep
	@test -e $(CONFIG) || (echo "please create a config.yaml first. Hint: check the config.example.yaml" && exit 1)
	go run ./cmd/slack-bot-builder -config $(CONFIG) -core . -workdir build/run -tags pprof -output build/slack-bot-run$(GOEXE)
	./build/slack-bot-run$(GOEXE) -config $(CONFIG)

# chat with the bot in the terminal, including the plugins of the config
run-cli: dep
	@test -e $(CONFIG) || (echo "please create a config.yaml first. Hint: check the config.example.yaml" && exit 1)
	go run ./cmd/slack-bot-builder -config $(CONFIG) -core . -workdir build/run -output build/slack-bot-run$(GOEXE) -cli-output build/cli-run$(GOEXE)
	./build/cli-run$(GOEXE) -config $(CONFIG)

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
	$(call in-plugins,go fix ./... && golangci-lint run --fix)

docker-build:
	docker build . --force-rm -t brainexe/slack-bot:latest

docker-push:
	docker push brainexe/slack-bot:latest

test: dep
	go test ./...
	$(call in-plugins,go test ./...)

test-race: dep
	go test ./... -race
	$(call in-plugins,go test ./... -race)

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
