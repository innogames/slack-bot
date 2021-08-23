package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/innogames/slack-bot/v2/bot"
	log "github.com/sirupsen/logrus"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/bot/tester"
	"github.com/innogames/slack-bot/v2/bot/util"
)

// starts a interactive shell to communicate with a mocked slack server and execute real commands
func main() {
	configFile := flag.String("config", "", "Path to config.yaml. Can be a directory which will load all '*.yaml' inside")
	verbose := flag.Bool("verbose", false, "More verbose output")
	flag.Parse()

	color.Info.Println("Hey! I'm your Slack Emulator. Call 'help' to get a list of all supported commands")

	var cfg config.Config
	var err error
	if *configFile == "" {
		fmt.Println("Hint: You can pass a custom config file by using '-config config.yaml'")
		cfg = config.DefaultConfig
	} else {
		cfg, err = config.Load(*configFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	cfg.Slack = config.Slack{}
	cfg.Slack.SocketToken = "xapp-something"

	cfg.Logger.Level = "error"
	if *verbose {
		cfg.Logger.Level = "debug"
	}

	cfg.AdminUsers = config.UserList{
		"cli",
	}

	bot.InitLogger(cfg.Logger)

	ctx := util.NewServerContext()

	startCli(ctx, os.Stdin, os.Stdout, cfg)
}

func startCli(ctx *util.ServerContext, input io.Reader, output io.Writer, cfg config.Config) {
	// set an empty storage -> just store data in Ram
	_ = storage.InitStorage("")

	// starts a local http server which is mocking the needed Slack API
	fakeSlack := tester.StartFakeSlack(&cfg, output)
	defer fakeSlack.Stop()

	realBot := tester.StartBot(cfg)

	color.SetOutput(output)
	color.Note.Print("Type in your command:\n")
	reader := bufio.NewReader(input)

	// loop to send stdin input to slack bot
	go func() {
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				continue
			}

			tester.HandleMessage(text)
		}
	}()

	realBot.Run(ctx)
}
