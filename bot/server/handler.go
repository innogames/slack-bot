package server

import (
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"net/http"
)

func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("I'm your friendly slack-bot"))
}

// see https://api.slack.com/authentication/verifying-requests-from-slack
func (s *Server) verifyRequest(w http.ResponseWriter, r *http.Request, body []byte) bool {
	// verifyRequest signature
	sv, err := slack.NewSecretsVerifier(r.Header, s.cfg.SigningSecret)
	if err != nil {
		s.error(w, err, http.StatusUnauthorized)
		return false
	}

	sv.Write(body)
	if err := sv.Ensure(); err != nil {
		s.error(w, err, http.StatusUnauthorized)
		return false
	}

	return true
}

func (s *Server) error(w http.ResponseWriter, err error, status int) {
	log.Error(err)
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
