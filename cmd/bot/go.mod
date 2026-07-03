module github.com/innogames/slack-bot/v2/cmd/bot

go 1.25.0

// the default "batteries included" bot binary: framework + all in-repo plugins.
// For a lean bot or a custom plugin set, create an own main package like this one
// and only import the plugins you need - see docs/plugins.md
// this module is only buildable within the repository, where the go.work
// workspace resolves the slack-bot + plugin modules to the local directories
require (
	github.com/innogames/slack-bot/v2 v2.3.17
	github.com/innogames/slack-bot/v2/plugins/aws v0.0.0
	github.com/innogames/slack-bot/v2/plugins/ripeatlas v0.0.0
)
