package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/vcs"
	"github.com/innogames/slack-bot/command"
	"github.com/sirupsen/logrus"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config.yaml. Can be a glob pattern like 'config/*.yaml'")
	flag.Parse()

	cfg, err := config.LoadPattern(*configFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger := bot.GetLogger(cfg)
	logger.Infof("Loaded config from %s", *configFile)

	_, err = storage.InitStorage(cfg.StoragePath)
	checkError(err, logger)

	slackClient := client.GetSlackClient(cfg.Slack, logger)

	vcs.InitBranchWatcher(cfg, logger)

	// make sure we're random enough
	rand.Seed(time.Now().UnixNano())

	// set default timezone
	time.Local, err = time.LoadLocation(cfg.Timezone)
	checkError(err, logger)

	commands := command.GetCommands(slackClient, cfg, logger)

	b := bot.NewBot(cfg, slackClient, logger, commands)
	err = b.Init()
	checkError(err, logger)

	// clean shutdown on sigterm/sigint
	var stopChan = make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// start main loop!
	b.HandleMessages(stopChan)
}

func checkError(err error, logger *logrus.Logger) {
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
