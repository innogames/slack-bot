package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
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

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	configFile := flag.String("config", "config.yaml", "Path to config.yaml. Can be a directory which will load all '*.yaml' inside")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	logger := bot.GetLogger(cfg.Logger)
	logger.Infof("Loaded config from %s", *configFile)

	err = storage.InitStorage(cfg.StoragePath)
	checkError(err, logger)

	slackClient := client.GetSlackClient(cfg.Slack, logger)

	vcs.InitBranchWatcher(cfg, logger)

	// todo(matze) check if we really want it here
	// make sure we're random enough
	rand.Seed(time.Now().UnixNano())

	// set global default timezone
	if cfg.Timezone != "" {
		time.Local, err = time.LoadLocation(cfg.Timezone)
		checkError(err, logger)
	}

	commands := command.GetCommands(slackClient, cfg, logger)

	b := bot.NewBot(cfg, slackClient, logger, commands)
	err = b.Init()
	checkError(err, logger)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	// start main loop!
	go b.HandleMessages(ctx, wg)

	var stopChan = make(chan os.Signal, 2)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-stopChan
	logger.Infof("Starting shutdown")
	cancel()
	wg.Wait()
	logger.Infof("Shutdown done, bye bye!")
}

func checkError(err error, logger *logrus.Logger) {
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
