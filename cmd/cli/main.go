package main

import (
	"bufio"
	"flag"
	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/tester"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"io"
	"os"
	"strings"
)

// starts a interactive shell to communicate with a mocked slack server and execute real commands
func main() {
	var verbose bool
	flag.BoolVar(&verbose, "verbose", false, "More verbose output")

	// todo add path to config
	flag.Parse()

	cfg := config.Config{}
	if verbose {
		cfg.Logger.Level = "debug"
	}

	ctx := util.NewServerContext()

	startCli(ctx, os.Stdin, os.Stdout, cfg)
}

func startCli(ctx *util.ServerContext, input io.Reader, output io.Writer, cfg config.Config) {
	ctx.RegisterChild()
	defer ctx.ChildDone()

	// set an empty storage -> just store data in Ram
	storage.InitStorage("")

	// starts a local http server which is mocking the needed Slack API
	fakeSlack := tester.StartFakeSlack(&cfg, output)
	defer fakeSlack.Stop()

	realBot := tester.StartBot(cfg)
	go realBot.ListenForMessages(ctx)

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

			client.HandleMessageWithDoneHandler(message).Wait()
		}
	}
}
