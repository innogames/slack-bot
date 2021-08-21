package main

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load("config.yaml")
	checkError(err)

	bot.InitLogger(cfg.Logger)

	slackClient, err := client.GetSlackClient(cfg.Slack)
	checkError(err)

	// get the list of all default commands
	commands := command.GetCommands(slackClient, cfg)

	// and add our custom one
	commands.AddCommand(exampleCommand{slackClient})

	b := bot.NewBot(cfg, slackClient, commands)
	err = b.Init()
	checkError(err)

	// start main loop!
	ctx := util.NewServerContext()
	b.Run(ctx)
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
