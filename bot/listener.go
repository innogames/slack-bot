package bot

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Run is blocking method to handle new incoming events...from different sources
func (b *Bot) Run(ctx *util.ServerContext) {
	b.startRunnables(ctx)

	// initialize Socket Mode:
	// https://api.slack.com/apis/connections/socket
	go func() {
		if err := b.slackClient.Socket.Run(); err != nil {
			log.Errorf("Socket run error: %v", err)
		}
	}()

	// graceful shutdown via sigterm/sigint
	stopChan := make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case event := <-b.slackClient.Socket.Events:
			// message from Socket Mode
			b.handleSocketModeEvent(event)
		case message := <-client.InternalMessages:
			// e.g. triggered by "delay" or "macro" command. They are still executed in original event context
			// -> will post in same channel as the user posted the original command
			message.InternalMessage = true
			go b.ProcessMessage(message, false)
		case <-stopChan:
			// wait until other services are properly shut down
			ctx.StopTheWorld()
			return
		case <-ctx.Done():
			return
		}
	}
}

// startRunnables starts all background tasks and ctx.StopTheWorld() will stop them then properly
func (b *Bot) startRunnables(ctx *util.ServerContext) {
	// each command can have a background task which is executed in the background
	for _, cmd := range b.commands.commands {
		if runnable, ok := cmd.(Runnable); ok {
			go runnable.RunAsync(ctx)
		}
	}

	// special handler which are executed in the background
	stats.InitMetrics(b.config, ctx)
}

func (b *Bot) handleSocketModeEvent(event socketmode.Event) {
	if event.Request != nil && event.Type != socketmode.EventTypeHello {
		b.slackClient.Socket.Ack(*event.Request)
	}

	switch event.Type {
	case socketmode.EventTypeConnectionError, socketmode.EventTypeErrorBadMessage, socketmode.EventTypeErrorWriteFailed, socketmode.EventTypeIncomingError, socketmode.EventTypeInvalidAuth:
		log.Warnf("Socket Mode error: %s - %s", event.Type, event.Data)
	case socketmode.EventTypeEventsAPI:
		b.handleEvent(event.Data.(slackevents.EventsAPIEvent))
	case socketmode.EventTypeInteractive:
		b.handleInteraction(event.Data.(slack.InteractionCallback))
	case socketmode.EventTypeConnected, socketmode.EventTypeConnecting, socketmode.EventTypeHello, socketmode.EventTypeDisconnect, socketmode.EventTypeSlashCommand:
		// ignore
	default:
		log.Infof("Unexpected event type received: %s\n", event.Type)
	}
}
