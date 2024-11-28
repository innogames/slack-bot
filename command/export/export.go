package export

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

// limit the number of exported messages (incl thread messages)
const limit = 3000

// NewExportCommand is a constructor to create a new export command
func NewExportCommand(base bot.BaseCommand) bot.Command {
	return &exportCommand{base}
}

type exportCommand struct {
	bot.BaseCommand
}

func (c *exportCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`export channel #?(?P<channel>[\w\-]+) as csv`, c.exportChannel),
	)
}

func (c *exportCommand) exportChannel(match matcher.Result, message msg.Message) {
	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	channelID, channelName := client.GetChannelIDAndName(match.GetString("channel"))

	buffer, lines, err := exportChannelMessagesToBuffer(c.SlackClient, channelID)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	if lines >= limit {
		c.SendMessage(message, fmt.Sprintf("Attention: The export is limited to %d messages.", limit))
	}
	_, err = c.UploadFile(slack.UploadFileV2Parameters{
		Filename:        "export.csv",
		FileSize:        buffer.Len(),
		Channel:         message.Channel,
		ThreadTimestamp: message.Timestamp,
		Reader:          buffer,
		InitialComment:  fmt.Sprintf("Export %d messages from channel <#%s|%s>", lines, channelID, channelName),
	})
	if err != nil {
		c.ReplyError(message, err)
	}
}

func (c *exportCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "export channel <name> as csv",
			Description: "export the given channel as csv",
		},
	}
}

func getConversations(client client.SlackClient, channelID string) ([]slack.Message, error) {
	var messages []slack.Message
	params := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     1000,
	}
	for {
		history, err := client.GetConversationHistory(&params)
		if err != nil {
			return nil, err
		}
		messages = append(messages, history.Messages...)
		if !history.HasMore {
			break
		}
		params.Cursor = history.ResponseMetaData.NextCursor
		time.Sleep(500 * time.Millisecond) // Avoid rate limiting
	}
	return messages, nil
}

func exportChannelMessagesToBuffer(client client.SlackClient, channelID string) (*bytes.Buffer, int, error) {
	buffer := new(bytes.Buffer)
	messages, err := getConversations(client, channelID)
	if err != nil {
		return nil, 0, err
	}

	writer := csv.NewWriter(buffer)

	// Write the CSV header
	_ = writer.Write([]string{"timestamp", "thread", "user-id", "text"})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp > messages[j].Timestamp
	})

	lines := 0
	for _, message := range messages {
		if message.SubType == "channel_join" || message.Text == "" {
			continue
		}

		// main messages from the channel
		writeLine(message, "", writer)
		lines++

		// write the thread messages (if any)
		if message.ThreadTimestamp != "" && message.ThreadTimestamp == message.Timestamp {
			threadTS := message.ThreadTimestamp
			threadMessages, err := client.GetThreadMessages(msg.MessageRef{
				Channel: channelID,
				Thread:  threadTS,
			})
			if err != nil {
				return nil, lines, err
			}
			for _, threadMessage := range threadMessages {
				// ignore first message, as we have it in the main channel already
				if threadMessage.Timestamp != threadTS {
					writeLine(threadMessage, threadTS, writer)
					lines++
				}
			}
			// rate limit...
			time.Sleep(200 * time.Millisecond)
		}

		if lines > limit {
			break
		}
	}

	writer.Flush()
	return buffer, lines, nil
}

func writeLine(message slack.Message, threadTimestamp string, writer *csv.Writer) {
	if message.Text == "" {
		return
	}

	user := message.User
	if message.User == "" {
		user = message.BotID
	}

	_ = writer.Write([]string{
		message.Timestamp,
		threadTimestamp,
		user,
		message.Text,
	})
}
