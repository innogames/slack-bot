package jenkins

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/mocks"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/stretchr/testify/assert"
)

func TestInformIdle(t *testing.T) {
	slackClient, jenkins, base := getTestJenkinsCommand()

	trigger := newIdleWatcherCommand(base).(*idleWatcherCommand)

	command := bot.Commands{}
	command.AddCommand(trigger)

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	t.Run("Test busy job", func(t *testing.T) {
		assert.Equal(t, 0, queue.CountCurrentJobs())

		message := msg.Message{}
		message.Text = "wait until jenkins is idle"

		// we don't have time to wait some minutes for the second check...
		trigger.checkInterval = time.Millisecond * 1

		ctx := context.Background()
		// first call: ne job is idle...the other one is still busy
		jenkins.On("GetAllNodes", ctx).Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1, "swarm1"),
			getNodeWithExecutors(1, 1, "swarm2"),
		}, nil).Once()

		// second call: idle!
		jenkins.On("GetAllNodes", ctx).Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1, "swarm1"),
			getNodeWithExecutors(0, 2, "swarm2"),
		}, nil).Once()

		mocks.AssertSlackMessage(slackClient, message, "There are 1 builds running...")
		mocks.AssertReaction(slackClient, waitingReaction, message)
		mocks.AssertRemoveReaction(slackClient, waitingReaction, message)
		mocks.AssertSlackMessage(slackClient, message, "No job is running anymore")
		mocks.AssertReaction(slackClient, doneReaction, message)

		actual := command.Run(message)
		assert.True(t, actual)

		// wait until watcher is ready
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, 0, queue.CountCurrentJobs())
	})

	t.Run("Test no one job running", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "wait until jenkins is idle"

		// we have 2 nodes with only idle executors
		mocks.AssertReaction(slackClient, doneReaction, message)

		ctx := context.Background()
		jenkins.On("GetAllNodes", ctx).Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1, "swarm1"),
			getNodeWithExecutors(0, 2, "swarm2"),
		}, nil)
		mocks.AssertSlackMessage(slackClient, message, "There are no jobs running right now!")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test job running on other node", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "wait until jenkins node swarm1 is idle"

		// we have 2 nodes with only idle executors
		mocks.AssertReaction(slackClient, doneReaction, message)

		ctx := context.Background()
		jenkins.On("GetAllNodes", ctx).Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1, "swarm1"),
			getNodeWithExecutors(1, 2, "swarm2"),
		}, nil)
		mocks.AssertSlackMessage(slackClient, message, "There are no jobs running right now!")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}

// some hacky way to mock some gojenkins.Node with the expected number of executors
func getNodeWithExecutors(running int, idle int, name string) *gojenkins.Node {
	node := &gojenkins.Node{
		Raw: &gojenkins.NodeResponse{
			DisplayName: name,
		},
	}

	executors := make([]executor, 0, running+idle)
	for i := 0; i < idle; i++ {
		executors = append(executors, executor{
			currentExecutable{0},
		})
	}
	for i := 0; i < running; i++ {
		executors = append(executors, executor{
			currentExecutable{12},
		})
	}
	executorJSON, _ := json.Marshal(executors)
	json.Unmarshal(executorJSON, &node.Raw.Executors)

	return node
}

// structs to mock the node...
type executor struct {
	CurrentExecutable currentExecutable `json:"CurrentExecutable"`
}

type currentExecutable struct {
	Number int `json:"number"`
}
