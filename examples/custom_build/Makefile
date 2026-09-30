# slack-bot version which is used to build the bot, pin it for reproducible builds, e.g. "make build BOT_VERSION=v2.5.0"
BOT_VERSION ?= latest
BUILDER ?= go run github.com/innogames/slack-bot/v2/cmd/slack-bot-builder@$(BOT_VERSION)

.PHONY: build run run-cli docker-build

# builds the bot and the cli binary including all plugins of the config.yaml
build:
	$(BUILDER) -config config.yaml -output build/slack-bot -cli-output build/cli

run: build
	./build/slack-bot -config config.yaml

# chat with the bot in the terminal, without a Slack connection
run-cli: build
	./build/cli -config config.yaml

docker-build:
	docker build . --build-arg BOT_VERSION=$(BOT_VERSION) -t my-slack-bot
