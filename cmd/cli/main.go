package main

import (
	"bufio"
	"context"
	"flag"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/tester"
	"github.com/sirupsen/logrus"
)

// starts a interactive shell to communicate with a mocked slack server and execute real commands
func main() {
	var verbose bool

	flag.BoolVar(&verbose, "v", false, "-v to use verbose logging")
	flag.Parse()

	cfg := config.Config{}
	ctx, _ := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	startCli(ctx, wg, os.Stdin, os.Stdout, cfg, verbose)
}

func startCli(ctx context.Context, wg *sync.WaitGroup, input io.Reader, output io.Writer, cfg config.Config, verbose bool) {
	wg.Add(1)
	defer wg.Done()

	var logger *logrus.Logger
	if verbose {
		logger = bot.GetLogger(cfg.Logger)
	} else {
		logger = tester.GetNullLogger()
	}

	// set an empty storage -> just store data in Ram
	storage.InitStorage("")

	// starts a local http server which is mocking the needed Slack API
	fakeSlack := tester.StartFakeSlack(&cfg, output)
	defer fakeSlack.Stop()

	realBot := tester.StartBot(cfg, logger)
	go realBot.HandleMessages(ctx, wg)

	color.SetOutput(output)
	color.Red.Print("Type in your command:\n")
	reader := bufio.NewReader(input)

	// loop to send stdin input to slack bot
	for {
		select {
		case <-ctx.Done():
			return
		default:
			text, err := reader.ReadString('\n')
			if err != nil {
				continue
			}

			color.Blue.Printf(">>>> %s\n", strings.TrimSuffix(text, "\n"))

			message := msg.Message{}
			message.Text = text
			message.Channel = tester.TestChannel
			message.User = "cli"

			client.InternalMessages <- message
		}
	}
}
