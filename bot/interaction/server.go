package interaction

import (
	"context"
	"net/http"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

func NewServer(cfg config.Server, logger *log.Logger, slackClient *client.Slack, allowedUsers map[string]string) *Server {
	return &Server{cfg: cfg, logger: logger, slackClient: slackClient, allowedUsers: allowedUsers}
}

type Server struct {
	cfg          config.Server
	logger       *log.Logger
	server       *http.Server
	slackClient  *client.Slack
	allowedUsers map[string]string
}

// https://api.slack.com/messaging/interactivity
func (s *Server) StartServer() {
	http.HandleFunc("/health", s.healthCheckHandler)
	http.HandleFunc("/commands", s.interactionHandler)

	s.server = &http.Server{
		Addr: s.cfg.Listen,
	}

	s.logger.Infof("Started Server on %s", s.cfg.Listen)

	err := s.server.ListenAndServe()
	if err != nil {
		s.logger.Fatal(err)
	}
}

func (s *Server) Stop() error {
	s.logger.Infof("Shutting down server")

	return s.server.Shutdown(context.Background())
}

func getChangedMessage(newMessage slack.Message, actionId string) slack.Message {
	for _, blocks := range newMessage.Blocks.BlockSet {
		if actionBlock, ok := blocks.(*slack.ActionBlock); ok {
			for _, block := range actionBlock.Elements.ElementSet {
				if buttonBlock, ok := block.(*slack.ButtonBlockElement); ok {
					if buttonBlock.Value == actionId {
						buttonBlock.Value = ""
						buttonBlock.Text.Text = buttonBlock.Text.Text + " (already clicked)"
					}
				}
			}
		}
	}

	return newMessage
}
