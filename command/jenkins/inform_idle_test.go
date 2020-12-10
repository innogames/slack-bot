package jenkins

import (
	"encoding/json"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestInformIdle(t *testing.T) {
	slackClient, jenkins, base := getTestJenkinsCommand()

	trigger := newIdleWatcherCommand(base).(*idleWatcherCommand)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test busy job", func(t *testing.T) {
		assert.Equal(t, 0, queue.CountCurrentJobs())

		message := msg.Message{}
		message.Text = "wait until jenkins is idle"

		// we don't have time to wait some minutes for the second check...
		trigger.checkInterval = time.Millisecond * 1

		// first call: ne job is idle...the other one is still busy
		jenkins.On("GetAllNodes").Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1),
			getNodeWithExecutors(1, 1),
		}, nil).Once()

		// second call: idle!
		jenkins.On("GetAllNodes").Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1),
			getNodeWithExecutors(0, 2),
		}, nil).Once()

		slackClient.On("AddReaction", waitingReaction, message)
		slackClient.On("RemoveReaction", waitingReaction, message)
		slackClient.On("SendMessage", message, "No job is running anymore").Return("")
		slackClient.On("AddReaction", doneReaction, message)

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
		jenkins.On("GetAllNodes").Return([]*gojenkins.Node{
			getNodeWithExecutors(0, 1),
			getNodeWithExecutors(0, 2),
		}, nil)
		slackClient.On("SendMessage", message, "There are no jobs running right now!").Return("")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.NotNil(t, help)
	})
}

// some hacky way to mock some gojenkins.Node with the expected number of executors
func getNodeWithExecutors(running int, idle int) *gojenkins.Node {
	node := &gojenkins.Node{
		Raw: &gojenkins.NodeResponse{},
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
