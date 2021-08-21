package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command"

	log "github.com/sirupsen/logrus"
)

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	var configFile string
	var verbose bool
	var showConfig bool
	flag.StringVar(&configFile, "config", "config.yaml", "Path to config.yaml. Can be a directory which will load all '*.yaml' inside")
	flag.BoolVar(&verbose, "verbose", false, "More verbose output")
	flag.BoolVar(&showConfig, "show-config", false, "Print the config as JSON and exit")
	flag.Parse()

	cfg, err := config.Load(configFile)
	checkError(err)

	if verbose {
		cfg.Logger.Level = "debug"
	}

	if showConfig {
		fmt.Println(config.Dump(cfg))
		os.Exit(0)
	}

	bot.InitLogger(cfg.Logger)
	log.Infof("Loaded config from %s", configFile)

	err = storage.InitStorage(cfg.StoragePath)
	checkError(err)

	slackClient, err := client.GetSlackClient(cfg.Slack)
	checkError(err)

	// get the list of all default commands
	commands := command.GetCommands(slackClient, cfg)

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
