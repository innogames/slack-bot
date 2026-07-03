package main

import (
	"github.com/innogames/slack-bot/v2/bot/app"

	// compile all in-repo plugins into the default bot binary,
	// see docs/plugins.md on how to cherry-pick plugins or add own ones
	_ "github.com/innogames/slack-bot/v2/plugins/aws"
	_ "github.com/innogames/slack-bot/v2/plugins/ripeatlas"
)

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	app.Run()
}
