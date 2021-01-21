package main

import (
	"flag"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/vcs"
	"github.com/innogames/slack-bot/command"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	var configFile string
	var verbose bool
	flag.StringVar(&configFile, "config", "config.yaml", "Path to config.yaml. Can be a directory which will load all '*.yaml' inside")
	flag.BoolVar(&verbose, "verbose", false, "More verbose output")
	flag.Parse()

	cfg, err := config.Load(configFile)
	checkError(err)

	if verbose {
		cfg.Logger.Level = "debug"
	}

	bot.InitLogger(cfg.Logger)
	log.Infof("Loaded config from %s", configFile)

	err = storage.InitStorage(cfg.StoragePath)
	checkError(err)

	slackClient := client.GetSlackClient(cfg.Slack)

	ctx := util.NewServerContext()
	go vcs.InitBranchWatcher(&cfg, ctx) // todo move into some command to init branch watcher

	// set global default timezone
	if cfg.Timezone != "" {
		time.Local, err = time.LoadLocation(cfg.Timezone)
		checkError(err)
	}

	commands := command.GetCommands(slackClient, cfg)

	b := bot.NewBot(cfg, slackClient, commands)
	err = b.Init()
	checkError(err)

	// start main loop!
	go b.ListenForMessages(ctx)

	var stopChan = make(chan os.Signal, 2)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// listen for messages until we receive sigterm/sigint
	<-stopChan

	ctx.StopTheWorld()
	log.Infof("Shutdown done, bye bye!")
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
