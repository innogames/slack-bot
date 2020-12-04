package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/innogames/slack-bot/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNodes(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	command := bot.Commands{}
	command.AddCommand(newNodesCommand(base))

	t.Run("Test invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "nodes"

		actual := command.Run(message)
		assert.Equal(t, false, actual)
	})

	t.Run("Fetch with error", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jenkins nodes"

		jenkinsClient.On("GetAllNodes").Return(nil, fmt.Errorf("an error occurred"))
		slackClient.On("ReplyError", message, fmt.Errorf("an error occurred")).Return(true)
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})

	t.Run("Fetch nodes", func(t *testing.T) {
		jenkinsClient := &mocks.Client{}

		command := bot.Commands{}
		command.AddCommand(newNodesCommand(base))

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

		jenkinsClient.On("GetAllNodes").Return(nodes, nil)
		slackClient.On("SendMessage", message, "*2 Nodes*\n- *Node 1* - status: :check_mark: - executors: 0\n- *Node 2* - status: :red_circle: - executors: 0\n").Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
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
		command.AddCommand(newNodesCommand(base))

		message := msg.Message{}
		message.Text = "jenkins nodes"

		slackClient.On("SendMessage", message, mock.Anything).Return("")
		actual := command.Run(message)
		assert.Equal(t, true, actual)
	})
}
