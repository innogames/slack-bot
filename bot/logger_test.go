package bot

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogger(t *testing.T) {
	cfg := config.Config{}

	logger := GetLogger(cfg.Logger)

	assert.IsType(t, &logrus.Logger{}, logger)
}
