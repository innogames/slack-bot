package main

import "github.com/innogames/slack-bot/v2/bot/app"
import _ "github.com/innogames/slack-bot/v2/plugins/test"

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	app.Run()
}
