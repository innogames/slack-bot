package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNodes(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	jenkins := &mocks.Client{}

	command := bot.Commands{}
	command.AddCommand(newNodesCommand(jenkins, slackClient))

	t.Run("Test invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "nodes"

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Fetch with error", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "jenkins nodes"

		jenkins.On("GetAllNodes").Return(nil, fmt.Errorf("an error occurred"))
		slackClient.On("ReplyError", event, fmt.Errorf("an error occurred")).Return(true)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})

	t.Run("Fetch nodes", func(t *testing.T) {
		jenkins := &mocks.Client{}

		command := bot.Commands{}
		command.AddCommand(newNodesCommand(jenkins, slackClient))

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

		event := slack.MessageEvent{}
		event.Text = "jenkins nodes"

		jenkins.On("GetAllNodes").Return(nodes, nil)
		slackClient.On("Reply", event, "*2 Nodes*\n- *Node 1* - status: :check_mark: - executors: 0\n- *Node 2* - status: :red_circle: - executors: 0\n")
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}

func TestRealNodes(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	t.Run("Fetch real nodes", func(t *testing.T) {
		cfg := config.Jenkins{
			Host: "http://ci.jenkins-ci.org",
		}
		client, err := jenkins.GetClient(cfg)
		assert.Nil(t, err)

		command := bot.Commands{}
		command.AddCommand(newNodesCommand(client, slackClient))

		event := slack.MessageEvent{}
		event.Text = "jenkins nodes"

		slackClient.On("Reply", event, mock.Anything)
		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
