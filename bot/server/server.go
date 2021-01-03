package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// NewServer is used to receive slack interactions
func NewServer(cfg config.Server, slackClient client.SlackClient, messageHandler messageHandler) *Server {
	return &Server{cfg: cfg, slackClient: slackClient, messageHandler: messageHandler}
}

type messageHandler func(message *slack.MessageEvent)

type Server struct {
	cfg            config.Server
	messageHandler messageHandler
	server         http.Server
	slackClient    client.SlackClient
}

// StartServer to receive slack interactions or events via event-api
// https://api.slack.com/messaging/interactivity
// https://api.slack.com/events-api
func (s *Server) StartServer() {
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/health", s.healthCheckHandler)
	http.HandleFunc("/interactions", s.interactionHandler)
	http.HandleFunc("/events", s.eventHandler)

	s.server = http.Server{
		Addr: s.cfg.Listen,
	}

	log.Infof("Started Server on %s", s.cfg.Listen)

	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// Stop the http server to receive slack interactions
func (s *Server) Stop() error {
	return s.server.Shutdown(context.Background())
}
