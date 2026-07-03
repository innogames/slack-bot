module github.com/innogames/slack-bot/v2/cmd/cli

go 1.25.0

// local CLI tool to test the bot: framework + all in-repo plugins, like cmd/bot
// this module is only buildable within the repository, where the go.work
// workspace resolves the slack-bot + plugin modules to the local directories
require (
	github.com/gookit/color v1.6.0
	github.com/innogames/slack-bot/v2 v2.3.17
	github.com/innogames/slack-bot/v2/plugins/aws v0.0.0
	github.com/innogames/slack-bot/v2/plugins/ripeatlas v0.0.0
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.11.1
)
