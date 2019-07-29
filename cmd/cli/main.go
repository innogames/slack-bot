package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/storage"
	"github.com/innogames/slack-bot/bot/tester"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// starts a interactive shell to communicate with a fake slack server and execute real commands
func main() {
	var logger *logrus.Logger
	var verbose bool

	flag.BoolVar(&verbose, "v", false, "-v to use verbose logging")
	flag.Parse()

	cfg := config.Config{}

	if verbose {
		logger = bot.GetLogger(cfg)
	} else {
		logger = tester.GetNullLogger()
	}

	storage.InitStorage("./storage_cli")
	defer storage.DeleteAll()

	fakeSlack := tester.StartFakeSlack(&cfg)
	defer fakeSlack.Stop()

	realBot := tester.StartBot(cfg, logger)
	kill := make(chan os.Signal, 1)
	go realBot.HandleMessages(kill)

	fmt.Println("Type in your command:")
	reader := bufio.NewReader(os.Stdin)

	// loop to print received messages from websocket connection
	go func() {
		for m := range fakeSlack.SeenFeed {
			var message slack.MessageEvent
			err := json.Unmarshal([]byte(m), &message)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if message.Type == "typing" {
				continue
			}
			color.Yellow("<<<< %s\n", message.Text)
		}
	}()

	// loop to send stdin input to slack bot
	for {
		text, _ := reader.ReadString('\n')
		color.Blue(">>>> %s", strings.TrimSuffix(text, "\n"))

		fakeSlack.SendMessageToBot(tester.TestChannel, text)
	}
}
