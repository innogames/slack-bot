# Custom bot build
Example of a repository with an own bot setup: the `config.yaml` defines the [plugins](../../docs/plugins.md),
`make build` builds a bot binary including them via the `slack-bot-builder`.

- `make build`: builds `build/slack-bot` and `build/cli`
- `make run-cli`: chat with the bot in the terminal, without a Slack connection
- `make run`: starts the bot, the Slack tokens are passed via `BOT_SLACK_TOKEN` and `BOT_SLACK_SOCKET_TOKEN`
- `make docker-build`: builds a Docker image with the bot binary
- `plugins/hello`: an own plugin, which is part of this repository
