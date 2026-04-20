package jenkins

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestJenkinsTrigger(t *testing.T) {
	slackClient, jenkinsClient, base := getTestJenkinsCommand()

	cfg := config.JenkinsJobs{
		"TestJob": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
			Trigger: "start test job",
		},
		"TestJobWithTrigger": {
			Parameters: []config.JobParameter{},
			Trigger:    "just do it",
		},
		"TestJobWithoutTrigger": {
			Parameters: []config.JobParameter{},
		},
		"Prefix/Test": {
			Parameters: []config.JobParameter{
				{Name: "PARAM1"},
			},
		},
	}

	trigger := newTriggerCommand(base, cfg, 5*time.Minute)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Len(t, help, 3)
	})

	t.Run("Trigger not existing job", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job NotExisting"

		expectedJobs := "Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger"
		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *NotExisting* is not startable. Possible jobs: \n - *"+expectedJobs+"*")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Trigger URL-encoded job name not whitelisted", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job backend/game-backend/xandreas%2FHC-7231"

		// The decoded job name should be used in the error message
		decodedJobName := "backend/game-backend/xandreas/HC-7231"
		expectedJobs := "Prefix/Test* \n - *TestJob* \n - *TestJobWithTrigger* \n - *TestJobWithoutTrigger"
		mocks.AssertSlackMessage(slackClient, message, "Sorry, job *"+decodedJobName+"* is not startable. Possible jobs: \n - *"+expectedJobs+"*")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Trigger URL-encoded job name that is whitelisted", func(t *testing.T) {
		// Create a separate config for this test to avoid affecting other tests
		testCfg := config.JenkinsJobs{
			"TestJob": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
				Trigger: "start test job",
			},
			"TestJobWithTrigger": {
				Parameters: []config.JobParameter{},
				Trigger:    "just do it",
			},
			"TestJobWithoutTrigger": {
				Parameters: []config.JobParameter{},
			},
			"Prefix/Test": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
			},
			"backend/game-backend/xandreas%2FHC-7231": {
				Parameters: []config.JobParameter{
					{Name: "PARAM1"},
				},
			},
		}

		testTrigger := newTriggerCommand(base, testCfg, 5*time.Minute)
		testCommand := bot.Commands{}
		testCommand.AddCommand(testTrigger)

		message := msg.Message{}
		message.Text = "trigger job backend%2Fgame-backend%2Fxandreas%252FHC-7231"

		// Expect an error about missing parameters since we're not providing PARAM1
		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")

		actual := testCommand.Run(message)
		assert.True(t, actual)
	})

	t.Run("Not enough parameters", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job TestJob"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start test job"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("matched prefixed trigger, but not all parameters provided", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job Prefix/Test"

		mocks.AssertError(slackClient, message, "sorry, you have to pass 1 parameters (PARAM1)")
		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("generic trigger", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job TestJob foo"

		mocks.AssertReaction(slackClient, "coffee", message)

		slackClient.On(
			"ReplyError",
			message,
			mock.Anything,
		)

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "TestJob").Return(nil, errors.New("404"))
		actual := command.Run(message)

		assert.True(t, actual)
	})

	t.Run("custom trigger", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "just do it"

		mocks.AssertReaction(slackClient, "coffee", message)

		slackClient.On(
			"ReplyError",
			message,
			mock.Anything,
		)

		ctx := context.Background()
		jenkinsClient.On("GetJob", ctx, "TestJobWithTrigger").Return(nil, errors.New("404"))
		actual := command.Run(message)

		assert.True(t, actual)
	})

	t.Run("No trigger found...do nothing", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "start foo job"

		actual := command.Run(message)
		assert.False(t, actual)
	})
}

func TestJenkinsTriggerApproval(t *testing.T) {
	slackClient, _, base := getTestJenkinsCommand()

	cfg := config.JenkinsJobs{
		"ApprovalJob": {
			Parameters: []config.JobParameter{
				{Name: "BRANCH", Default: "master"},
			},
			NeedsApproval: true,
		},
		"NormalJob": {
			Parameters: []config.JobParameter{
				{Name: "BRANCH", Default: "master"},
			},
		},
	}

	trigger := newTriggerCommand(base, cfg, 5*time.Minute)

	command := bot.Commands{}
	command.AddCommand(trigger)

	t.Run("trigger job with NeedsApproval sends DM instead of starting", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "trigger job ApprovalJob master"
		message.User = "U12345"

		// expect DM with blocks to user
		slackClient.On("SendBlockMessageToUser", "U12345", mock.AnythingOfType("[]slack.Block")).Return("dm-ts").Once()
		// expect channel notification
		mocks.AssertSlackMessage(slackClient, message, "Job *ApprovalJob* requires approval. Please check your direct messages.")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("approve non-existent approval", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jenkins approve doesnotexist"

		mocks.AssertSlackMessage(slackClient, message, "Approval not found or expired. Please re-trigger the job.")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("reject non-existent approval", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "jenkins reject doesnotexist"

		mocks.AssertSlackMessage(slackClient, message, "Approval not found or already handled.")

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("approve valid approval triggers job", func(t *testing.T) {
		slackClient2, jenkinsClient2, base2 := getTestJenkinsCommand()

		cfg2 := config.JenkinsJobs{
			"DeployProd": {
				Parameters: []config.JobParameter{
					{Name: "BRANCH", Default: "master"},
				},
				NeedsApproval: true,
			},
		}
		trigger2 := newTriggerCommand(base2, cfg2, 5*time.Minute).(*triggerCommand)
		cmd2 := bot.Commands{}
		cmd2.AddCommand(trigger2)

		originalMessage := msg.Message{}
		originalMessage.Text = "trigger job DeployProd master"
		originalMessage.User = "U12345"
		originalMessage.Channel = "C12345"

		// request approval first
		slackClient2.On("SendBlockMessageToUser", "U12345", mock.AnythingOfType("[]slack.Block")).Return("dm-ts").Once()
		mocks.AssertSlackMessage(slackClient2, originalMessage, "Job *DeployProd* requires approval. Please check your direct messages.")

		cmd2.Run(originalMessage)

		// find the stored approval ID
		trigger2.approvals.mu.Lock()
		var approvalID string
		for id := range trigger2.approvals.pending {
			approvalID = id
		}
		trigger2.approvals.mu.Unlock()
		assert.NotEmpty(t, approvalID)

		// now approve it
		approveMessage := msg.Message{}
		approveMessage.Text = "jenkins approve " + approvalID

		mocks.AssertSlackMessage(slackClient2, approveMessage, "Job *DeployProd* approved, starting build...")

		// expect the job to be triggered (will hit Jenkins client)
		mocks.AssertReaction(slackClient2, "coffee", originalMessage)
		slackClient2.On("ReplyError", originalMessage, mock.Anything)
		ctx := context.Background()
		jenkinsClient2.On("GetJob", ctx, "DeployProd").Return(nil, errors.New("404"))

		actual := cmd2.Run(approveMessage)
		assert.True(t, actual)
	})

	t.Run("reject valid approval", func(t *testing.T) {
		slackClient3, _, base3 := getTestJenkinsCommand()

		cfg3 := config.JenkinsJobs{
			"DeployProd": {
				Parameters: []config.JobParameter{
					{Name: "BRANCH", Default: "master"},
				},
				NeedsApproval: true,
			},
		}
		trigger3 := newTriggerCommand(base3, cfg3, 5*time.Minute).(*triggerCommand)
		cmd3 := bot.Commands{}
		cmd3.AddCommand(trigger3)

		originalMessage := msg.Message{}
		originalMessage.Text = "trigger job DeployProd master"
		originalMessage.User = "U12345"
		originalMessage.Channel = "C12345"

		// request approval
		slackClient3.On("SendBlockMessageToUser", "U12345", mock.AnythingOfType("[]slack.Block")).Return("dm-ts").Once()
		mocks.AssertSlackMessage(slackClient3, originalMessage, "Job *DeployProd* requires approval. Please check your direct messages.")
		cmd3.Run(originalMessage)

		// find the approval ID
		trigger3.approvals.mu.Lock()
		var approvalID string
		for id := range trigger3.approvals.pending {
			approvalID = id
		}
		trigger3.approvals.mu.Unlock()

		// reject it
		rejectMessage := msg.Message{}
		rejectMessage.Text = "jenkins reject " + approvalID

		mocks.AssertSlackMessage(slackClient3, rejectMessage, "Job *DeployProd* rejected.")
		mocks.AssertSlackMessage(slackClient3, originalMessage, "Job *DeployProd* was rejected.")

		actual := cmd3.Run(rejectMessage)
		assert.True(t, actual)
	})

	t.Run("approve expired approval", func(t *testing.T) {
		slackClient4, _, base4 := getTestJenkinsCommand()

		cfg4 := config.JenkinsJobs{
			"DeployProd": {
				Parameters:    []config.JobParameter{},
				NeedsApproval: true,
			},
		}
		trigger4 := newTriggerCommand(base4, cfg4, 1*time.Millisecond).(*triggerCommand)
		cmd4 := bot.Commands{}
		cmd4.AddCommand(trigger4)

		originalMessage := msg.Message{}
		originalMessage.Text = "trigger job DeployProd"
		originalMessage.User = "U12345"

		// request approval with very short timeout
		slackClient4.On("SendBlockMessageToUser", "U12345", mock.AnythingOfType("[]slack.Block")).Return("dm-ts").Once()
		mocks.AssertSlackMessage(slackClient4, originalMessage, "Job *DeployProd* requires approval. Please check your direct messages.")
		cmd4.Run(originalMessage)

		// find the approval ID
		trigger4.approvals.mu.Lock()
		var approvalID string
		for id := range trigger4.approvals.pending {
			approvalID = id
		}
		trigger4.approvals.mu.Unlock()

		// wait for expiry
		time.Sleep(5 * time.Millisecond)

		// try to approve
		approveMessage := msg.Message{}
		approveMessage.Text = "jenkins approve " + approvalID

		mocks.AssertSlackMessage(slackClient4, approveMessage, "Approval not found or expired. Please re-trigger the job.")

		actual := cmd4.Run(approveMessage)
		assert.True(t, actual)
	})
}
