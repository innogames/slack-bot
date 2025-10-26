package export

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExportCommandSimple(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	baseCommand := bot.BaseCommand{SlackClient: slackClient}
	command := NewExportCommand(baseCommand)

	t.Run("GetMatcher", func(t *testing.T) {
		matcher := command.GetMatcher()
		assert.NotNil(t, matcher)
	})

	t.Run("GetHelp", func(t *testing.T) {
		// Cast to the actual type to access GetHelp
		exportCmd := command.(*exportCommand)
		help := exportCmd.GetHelp()
		require.Len(t, help, 1)
		assert.Equal(t, "export channel <name> as csv", help[0].Command)
		assert.Equal(t, "export the given channel as csv", help[0].Description)
	})
}

func TestWriteLine(t *testing.T) {
	t.Run("write message line", func(t *testing.T) {
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)

		message := slack.Message{
			Msg: slack.Msg{
				Timestamp: "1234567890.123456",
				User:      "U1234567890",
				Text:      "Hello world",
			},
		}

		writeLine(message, "", writer)
		writer.Flush()

		content := buffer.String()
		assert.Contains(t, content, "1234567890.123456")
		assert.Contains(t, content, "U1234567890")
		assert.Contains(t, content, "Hello world")
	})

	t.Run("write thread message line", func(t *testing.T) {
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)

		message := slack.Message{
			Msg: slack.Msg{
				Timestamp: "1234567891.123456",
				User:      "U0987654321",
				Text:      "Thread reply",
			},
		}

		writeLine(message, "1234567890.123456", writer)
		writer.Flush()

		content := buffer.String()
		assert.Contains(t, content, "1234567891.123456")
		assert.Contains(t, content, "1234567890.123456") // thread timestamp
		assert.Contains(t, content, "U0987654321")
		assert.Contains(t, content, "Thread reply")
	})

	t.Run("skip empty message", func(t *testing.T) {
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)

		message := slack.Message{
			Msg: slack.Msg{
				Timestamp: "1234567890.123456",
				User:      "U1234567890",
				Text:      "",
			},
		}

		writeLine(message, "", writer)
		writer.Flush()

		content := buffer.String()
		assert.Empty(t, content)
	})

	t.Run("use bot ID when user is empty", func(t *testing.T) {
		var buffer bytes.Buffer
		writer := csv.NewWriter(&buffer)

		message := slack.Message{
			Msg: slack.Msg{
				Timestamp: "1234567890.123456",
				User:      "",
				BotID:     "B1234567890",
				Text:      "Bot message",
			},
		}

		writeLine(message, "", writer)
		writer.Flush()

		content := buffer.String()
		assert.Contains(t, content, "B1234567890")
		assert.Contains(t, content, "Bot message")
	})
}

func TestExportChannelMessagesToBufferSimple(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)

	t.Run("successful export to buffer", func(t *testing.T) {
		mockMessages := []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp: "1234567890.123456",
					User:      "U1234567890",
					Text:      "Hello world",
				},
			},
			{
				Msg: slack.Msg{
					Timestamp: "1234567891.123456",
					User:      "U0987654321",
					Text:      "Another message",
				},
			},
		}

		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		buffer, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Equal(t, 2, lines)
		assert.NotNil(t, buffer)

		// Verify CSV content
		content := buffer.String()
		assert.Contains(t, content, "timestamp,thread,user-id,text")
		assert.Contains(t, content, "Hello world")
		assert.Contains(t, content, "Another message")

		slackClient.AssertExpectations(t)
	})
}
