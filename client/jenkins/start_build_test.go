package jenkins

import (
	"encoding/json"
	"errors"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/msg"
	"testing"

	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
)

func TestStartBuild(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	cfg := config.JobConfig{}
	message := msg.Message{}

	t.Run("error fetching job", func(t *testing.T) {
		client := &mocks.Client{}
		jobName := "TestJob"
		params := map[string]string{}

		slackClient.On("AddReaction", iconPending, message)
		client.On("GetJob", jobName).Return(nil, errors.New("404"))

		err := TriggerJenkinsJob(cfg, jobName, params, slackClient, client, message)

		assert.Equal(t, "Job *TestJob* could not start job: 404", err.Error())
	})
}

func TestGetAttachment(t *testing.T) {
	jenkinsBuild := &gojenkins.Build{
		Raw: &gojenkins.BuildResponse{
			Result: gojenkins.STATUS_ABORTED,
			URL:    "https://jenkins.example.com/build/",
		},
	}
	message := "myMessage"
	actual := getAttachment(jenkinsBuild, message)
	jsonResponse, _ := json.Marshal(actual)

	expected := "{\"color\":\"#CCCCCC\",\"title\":\"myMessage\",\"title_link\":\"https://jenkins.example.com/build/\",\"text\":\"\",\"actions\":[{\"name\":\"\",\"text\":\"Build :black_circle_for_record:\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://jenkins.example.com/build/\"},{\"name\":\"\",\"text\":\"Console :page_with_curl:\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://jenkins.example.com/build/console\"},{\"name\":\"\",\"text\":\"Rebuild :arrows_counterclockwise:\",\"style\":\"default\",\"type\":\"button\",\"url\":\"https://jenkins.example.com/build/rebuild/parameterized\"}],\"blocks\":null}"
	assert.Equal(t, expected, string(jsonResponse))
}
