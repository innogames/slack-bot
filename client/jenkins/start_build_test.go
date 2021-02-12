package jenkins

import (
	"encoding/json"
	"errors"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStartBuild(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	cfg := config.JobConfig{}
	message := msg.Message{}

	t.Run("error fetching job", func(t *testing.T) {
		client := &mocks.Client{}
		jobName := "TestJob"
		params := map[string]string{}

		mocks.AssertReaction(slackClient, iconPending, message)
		client.On("GetJob", jobName).Return(nil, errors.New("404"))

		err := TriggerJenkinsJob(cfg, jobName, params, slackClient, client, message)

		assert.EqualError(t, err, "Job *TestJob* could not start job: 404")
	})
}

func TestGetAttachment(t *testing.T) {

	t.Run("Simple Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Result: gojenkins.STATUS_ABORTED,
				URL:    "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#CCCCCC","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :black_circle_for_record:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Rebuild :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/rebuild/parameterized"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})

	t.Run("Building Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Building: true,
				URL:      "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#E0E000","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Abort :bomb:","style":"danger","type":"button","url":"https://jenkins.example.com/build/stop/"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})
}
