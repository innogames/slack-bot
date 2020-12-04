package main

import (
	"flag"
	"github.com/innogames/slack-bot/bot/util"
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
	log "github.com/sirupsen/logrus"
	// comment in to profile live socket server via "/debug/pprof". e.g.:
	// attention: enable the server: section in the config!
	// https://golang.org/doc/diagnostics.html
	// curl localhost:4390/debug/pprof/heap\?debug=1 | less
	// curl localhost:4390/debug/pprof/allocs\?debug=1 | less
	// curl localhost:4390/debug/pprof/goroutine\?debug=1 | less
	// curl localhost:4390/debug/pprof/profile\?seconds=30 > /tmp/pprof.trace #
	// curl localhost:4390/debug/pprof/trace\?seconds=30 > /tmp/trace.trace #
	// _ "net/http/pprof"
)

// main entry point for the bot application. Listens on incoming slack messages and handles them
func main() {
	configFile := flag.String("config", "config.yaml", "Path to config.yaml. Can be a directory which will load all '*.yaml' inside")
	verbose := flag.Bool("verbose", false, "More verbose output")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	checkError(err)

	if *verbose {
		cfg.Logger.Level = "debug"
	}

	bot.InitLogger(cfg.Logger)
	log.Infof("Loaded config from %s", *configFile)

	err = storage.InitStorage(cfg.StoragePath)
	checkError(err)

	slackClient := client.GetSlackClient(cfg.Slack)

	ctx := util.NewServerContext()
	go vcs.InitBranchWatcher(&cfg, ctx)

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
	go b.HandleMessages(ctx)

	var stopChan = make(chan os.Signal, 2)

	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-stopChan

	ctx.StopTheWorld()
	log.Infof("Shutdown done, bye bye!")
}

func checkError(err error) {
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
