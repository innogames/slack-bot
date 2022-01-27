package jenkins

import (
	"context"
	"fmt"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client/jenkins"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		message.Text = "list jenkins nodes"

		ctx := context.TODO()
		jenkinsClient.On("GetAllNodes", ctx).Return(nil, fmt.Errorf("an error occurred")).Once()
		mocks.AssertError(slackClient, message, fmt.Errorf("an error occurred"))
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
			{
				Raw: &gojenkins.NodeResponse{
					DisplayName:        "Node 3",
					TemporarilyOffline: true,
				},
			},
		}

		message := msg.Message{}
		message.Text = "list jenkins nodes"

		ctx := context.Background()
		jenkinsClient.On("GetAllNodes", ctx).Return(nodes, nil).Once()
		mocks.AssertSlackMessage(slackClient, message, `*<https://jenkins.example.com/computer/|3 Nodes>*
• *<https://jenkins.example.com/computer/Node 1/|Node 1>* - online ✔ - busy executors: 0/0
• *<https://jenkins.example.com/computer/Node 2/|Node 2>* - offline 🔴 - busy executors: 0/0
• *<https://jenkins.example.com/computer/Node 3/|Node 3>* - temporary offline ⏸ - busy executors: 0/0

In total there are 0 build(s) running right now`)
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
		message.Text = "list jenkins nodes"

		slackClient.On("SendMessage", message, mock.Anything).Return("")
		actual := command.Run(message)
		assert.True(t, actual)
	})
}
