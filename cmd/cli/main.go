package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/tester"
	"github.com/sirupsen/logrus"
)

// starts a interactive shell to communicate with a fake slack server and execute real commands
func main() {
	kill := make(chan os.Signal, 1)

	startCli(os.Stdin, os.Stdout, kill)
}

func startCli(input io.Reader, output io.Writer, kill chan os.Signal) {
	var logger *logrus.Logger
	var verbose bool

	color.SetOutput(output)

	flag.BoolVar(&verbose, "v", false, "-v to use verbose logging")
	flag.Parse()

	cfg := config.Config{}

	if verbose {
		logger = bot.GetLogger(cfg)
	} else {
		logger = tester.GetNullLogger()
	}

	storage.InitStorage("")

	fakeSlack := tester.StartFakeSlack(&cfg)
	defer fakeSlack.Stop()

	realBot := tester.StartBot(cfg, logger)
	go realBot.HandleMessages(kill)

	fmt.Println("Type in your command:")
	reader := bufio.NewReader(input)

	// loop to send stdin input to slack bot
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			continue
		}

		color.Blue.Printf(">>>> %s", strings.TrimSuffix(text, "\n"))

		fakeSlack.SendMessageToBot(tester.TestChannel, text)
	}
}
