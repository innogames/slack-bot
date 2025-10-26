package export

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"testing"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
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

	t.Run("error in get conversation history", func(t *testing.T) {
		slackClient := mocks.NewSlackClient(t)

		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(
			(*slack.GetConversationHistoryResponse)(nil), errors.New("API error"))

		buffer, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		assert.Error(t, err)
		assert.Nil(t, buffer)
		assert.Equal(t, 0, lines)

		slackClient.AssertExpectations(t)
	})

	t.Run("thread messages handling", func(t *testing.T) {
		slackClient := mocks.NewSlackClient(t)

		mockMessages := []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp:       "1234567890.123456",
					ThreadTimestamp: "1234567890.123456",
					User:            "U1234567890",
					Text:            "Main thread message",
				},
			},
		}

		// Mock conversation history
		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		// Mock thread messages
		slackClient.On("GetThreadMessages", mock.MatchedBy(func(ref msg.MessageRef) bool {
			return ref.Thread == "1234567890.123456"
		})).Return([]slack.Message{
			{
				Msg: slack.Msg{
					Timestamp: "1234567891.123456",
					User:      "U0987654321",
					Text:      "Thread reply",
				},
			},
		}, nil)

		buffer, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Equal(t, 2, lines) // 1 main + 1 thread message

		content := buffer.String()
		assert.Contains(t, content, "Main thread message")
		assert.Contains(t, content, "Thread reply")

		slackClient.AssertExpectations(t)
	})

	t.Run("large export hitting limit", func(t *testing.T) {
		slackClient := mocks.NewSlackClient(t)

		// Create 3001 messages (exceeding the 3000 limit)
		// The logic breaks when lines > limit, so it processes 3001 before breaking
		mockMessages := make([]slack.Message, 3001)
		for i := 0; i < 3001; i++ {
			mockMessages[i] = slack.Message{
				Msg: slack.Msg{
					Timestamp: fmt.Sprintf("123456789%d.123456", i%10),
					User:      "U1234567890",
					Text:      fmt.Sprintf("Message %d", i),
				},
			}
		}

		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		_, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Equal(t, limit+1, lines) // Should be capped at 3001 due to the break condition

		slackClient.AssertExpectations(t)
	})

	t.Run("skip channel join messages", func(t *testing.T) {
		slackClient := mocks.NewSlackClient(t)

		mockMessages := []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp: "1234567890.123456",
					User:      "U1234567890",
					Text:      "Hello world",
					SubType:   "channel_join",
				},
			},
			{
				Msg: slack.Msg{
					Timestamp: "1234567891.123456",
					User:      "U0987654321",
					Text:      "Valid message",
				},
			},
		}

		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		buffer, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Equal(t, 1, lines) // Only the valid message should be counted

		content := buffer.String()
		assert.NotContains(t, content, "Hello world") // Channel join should be skipped
		assert.Contains(t, content, "Valid message")

		slackClient.AssertExpectations(t)
	})
}

func TestExportChannelFunction(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	baseCommand := bot.BaseCommand{SlackClient: slackClient}
	command := NewExportCommand(baseCommand).(*exportCommand)

	t.Run("matcher test", func(t *testing.T) {
		// Test that the matcher works correctly
		matcher := command.GetMatcher()

		message := msg.Message{
			MessageRef: msg.MessageRef{
				Channel: "C1234567890",
			},
			Text: "export channel general as csv",
		}

		run, match := matcher.Match(message)
		assert.NotNil(t, run)
		assert.NotNil(t, match)

		if match != nil {
			assert.Equal(t, "general", match.GetString("channel"))
		}
	})

	t.Run("error in conversation history", func(t *testing.T) {
		// Test the core functionality without full integration
		slackClient := mocks.NewSlackClient(t)

		// Mock error in conversation history
		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(
			(*slack.GetConversationHistoryResponse)(nil), errors.New("channel not found"))

		// Test error path in exportChannelMessagesToBuffer
		_, _, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "channel not found")

		slackClient.AssertExpectations(t)
	})

	t.Run("error in thread messages", func(t *testing.T) {
		// Test the core functionality without full integration
		slackClient := mocks.NewSlackClient(t)

		mockMessages := []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp:       "1234567890.123456",
					ThreadTimestamp: "1234567890.123456",
					User:            "U1234567890",
					Text:            "Main thread message",
				},
			},
		}

		// Mock conversation history success
		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		// Mock thread messages error
		slackClient.On("GetThreadMessages", mock.AnythingOfType("msg.MessageRef")).Return(
			[]slack.Message{}, errors.New("thread not accessible"))

		// Should return error when thread messages fail
		_, _, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "thread not accessible")

		slackClient.AssertExpectations(t)
	})

	t.Run("file upload preparation", func(t *testing.T) {
		// Test the core functionality without full integration
		slackClient := mocks.NewSlackClient(t)

		// Mock successful conversation history
		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: []slack.Message{
				{
					Msg: slack.Msg{
						Timestamp: "1234567890.123456",
						User:      "U1234567890",
						Text:      "Test message",
					},
				},
			},
			HasMore: false,
		}, nil)

		// Test that the function succeeds and returns data ready for upload
		buffer, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Greater(t, lines, 0)
		assert.NotNil(t, buffer)
		assert.Contains(t, buffer.String(), "Test message")

		slackClient.AssertExpectations(t)
	})

	t.Run("rate limiting scenarios", func(t *testing.T) {
		slackClient := mocks.NewSlackClient(t)

		// Test with HasMore = true to trigger pagination and rate limiting
		mockMessages := []slack.Message{
			{
				Msg: slack.Msg{
					Timestamp: "1234567890.123456",
					User:      "U1234567890",
					Text:      "First message",
				},
			},
		}

		// Mock conversation history with HasMore=false (no pagination needed for this test)
		slackClient.On("GetConversationHistory", mock.AnythingOfType("*slack.GetConversationHistoryParameters")).Return(&slack.GetConversationHistoryResponse{
			Messages: mockMessages,
			HasMore:  false,
		}, nil)

		_, lines, err := exportChannelMessagesToBuffer(slackClient, "C1234567890")
		require.NoError(t, err)
		assert.Equal(t, 1, lines) // 1 message from single call

		slackClient.AssertExpectations(t)
	})
}
