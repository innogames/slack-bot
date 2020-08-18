package jenkins

import (
	"errors"
	"testing"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func TestStartBuild(t *testing.T) {
	client := &mocks.Client{}
	slackClient := &mocks.SlackClient{}
	logger, _ := test.NewNullLogger()
	cfg := config.JobConfig{}
	event := slack.MessageEvent{}

	t.Run("error fetching job", func(t *testing.T) {
		jobName := "TestJob"
		params := map[string]string{}

		msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
		slackClient.On("AddReaction", iconPending, msgRef)
		client.On("GetJob", jobName).Return(nil, errors.New("404"))

		err := TriggerJenkinsJob(cfg, jobName, params, slackClient, client, event, logger)

		assert.Equal(t, "Job *TestJob* could not start job: 404", err.Error())
	})
}
