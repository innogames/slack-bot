package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNodes(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	cfg := config.Jenkins{}
	cfg.Host = "https://jenkins.example.com"

	command := bot.Commands{}
	command.AddCommand(newNodesCommand(base, cfg))

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "nodes"

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("Fetch with error", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jenkins nodes"

		jenkinsClient.On("GetAllNodes").Return(nil, fmt.Errorf("an error occurred")).Once()
		slackClient.On("ReplyError", message, fmt.Errorf("an error occurred")).Return(true)
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Fetch nodes", func(t *testing.T) {
		command := bot.Commands{}
		command.AddCommand(newNodesCommand(base, cfg))

		nodes := []*gojenkins.Node{
			{
				Raw: &gojenkins.NodeResponse{
					DisplayName: "Node 1",
					Offline:     false,
				},
			},
			{
				Raw: &gojenkins.NodeResponse{
					DisplayName: "Node 2",
					Offline:     true,
				},
			},
		}

		message := msg.Message{}
		message.Text = "jenkins nodes"

		jenkinsClient.On("GetAllNodes").Return(nodes, nil).Once()
		mocks.AssertSlackMessage(slackClient, message, `*<https://jenkins.example.com/computer/|2 Nodes>*
• *<https://jenkins.example.com/computer/Node 2/|Node 2>* - status: :red_circle: - busy executors: 0/0
• *<https://jenkins.example.com/computer/Node 1/|Node 1>* - status: :white_check_mark: - busy executors: 0/0
`)
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Equal(t, 1, len(help))
	})
}

// call a real jenkins server and check if the response is okay
func TestRealNodes(t *testing.T) {
	slackClient, _, base := getTestJenkinsCommand()

	t.Run("Fetch real nodes", func(t *testing.T) {
		cfg := config.Jenkins{
			Host: "http://ci.jenkins-ci.org",
		}
		client, err := jenkins.GetClient(cfg)
		assert.Nil(t, err)

		base.jenkins = client
		command := bot.Commands{}
		command.AddCommand(newNodesCommand(base, cfg))

		message := msg.Message{}
		message.Text = "jenkins nodes"

		slackClient.On("SendMessage", message, mock.Anything).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})
}
