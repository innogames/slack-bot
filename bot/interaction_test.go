package bot

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dummy command which set "called" flag when a command was called with the text "dummy"
type dummyCommand struct{}

func (d dummyCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("dummy", func(match matcher.Result, message msg.Message) {
	})
}

func TestInteraction(t *testing.T) {
	cfg := config.Config{}
	cfg.AllowedUsers = config.UserList{
		"user1",
	}

	rawSlackClient := &slack.Client{}
	slackClient := &client.Slack{Client: rawSlackClient}

	mockCommand := dummyCommand{}
	commands := &Commands{}
	commands.AddCommand(mockCommand)

	bot := NewBot(cfg, slackClient, commands)
	bot.auth = &slack.AuthTestResponse{
		UserID: "BOT",
	}
	bot.allowedUsers = config.UserMap{
		"user1": "user2",
	}

	t.Run("handle message event", func(t *testing.T) {
		message := &slackevents.MessageEvent{
			User:    "user1",
			Text:    "dummy",
			Channel: "1234", // we're not in a direct chang and have no annotation -> ignore the event
		}

		innerEvent := slackevents.EventsAPIInnerEvent{
			Data: message,
		}
		event := slackevents.EventsAPIEvent{
			Type:       slackevents.CallbackEvent,
			InnerEvent: innerEvent,
		}

		// we reset the stats count and expect later on that one command was processed completely
		stats.Set(stats.TotalCommands, 0)

		bot.handleEvent(event)
		time.Sleep(time.Millisecond * 20)

		commandsProcessed, err := stats.Get(stats.TotalCommands)
		require.NoError(t, err)
		assert.Equal(t, uint(0), commandsProcessed)
	})

	t.Run("don't handle edited message", func(t *testing.T) {
		message := &slackevents.MessageEvent{
			User:    "user1",
			Text:    "dummy",
			Channel: "D12",
			SubType: "message_changed",
		}

		innerEvent := slackevents.EventsAPIInnerEvent{
			Data: message,
		}
		event := slackevents.EventsAPIEvent{
			Type:       slackevents.CallbackEvent,
			InnerEvent: innerEvent,
		}

		// we reset the stats count and expect later on that one command was processed completely
		stats.Set(stats.TotalCommands, 0)

		bot.handleEvent(event)
		time.Sleep(time.Millisecond * 20)

		commandsProcessed, err := stats.Get(stats.TotalCommands)
		require.NoError(t, err)
		assert.Equal(t, uint(0), commandsProcessed)
	})

	t.Run("handle app mention event", func(t *testing.T) {
		message := &slackevents.AppMentionEvent{
			User:    "user1",
			Text:    "dummy",
			Channel: "D1234",
		}

		innerEvent := slackevents.EventsAPIInnerEvent{
			Data: message,
		}
		eventData := slackevents.EventsAPIEvent{
			Type:       slackevents.CallbackEvent,
			InnerEvent: innerEvent,
		}

		// we reset the stats count and expect later on that one command was processed completely
		stats.Set(stats.TotalCommands, 0)

		event := socketmode.Event{}
		event.Type = socketmode.EventTypeEventsAPI
		event.Data = eventData

		bot.handleSocketModeEvent(event)
		time.Sleep(time.Millisecond * 20)

		commandsProcessed, err := stats.Get(stats.TotalCommands)
		require.NoError(t, err)
		assert.Equal(t, uint(1), commandsProcessed)
	})

	t.Run("handle interaction", func(t *testing.T) {
		messageEvent := &slack.MessageEvent{}
		messageEvent.Channel = "D1234"
		messageEvent.User = "user1"

		action := slack.NewActionBlock("", client.GetInteractionButton("dummy", "my text", "dummy"))
		button := action.Elements.ElementSet[0].(*slack.ButtonBlockElement)
		actionID := button.Value
		assert.Equal(t, "dummy", actionID)

		callback := slack.InteractionCallback{
			Type: "block_actions",
			User: slack.User{
				ID: "user1",
			},
			ActionCallback: slack.ActionCallbacks{
				BlockActions: []*slack.BlockAction{
					{
						Value: actionID,
					},
				},
			},
		}

		success := bot.handleInteraction(callback)
		assert.True(t, success)

		// "press the button" again -> should not work!
		callback.ActionCallback.BlockActions[0].Value = ""
		callback.Type = "block_actions"
		success = bot.handleInteraction(callback)
		assert.False(t, success)
	})

	t.Run("handle invalid interaction", func(t *testing.T) {
		callback := slack.InteractionCallback{
			Type: "block_actions",
			User: slack.User{
				ID: "user1",
			},
			ActionCallback: slack.ActionCallbacks{
				BlockActions: []*slack.BlockAction{
					{
						Value: "",
					},
				},
			},
		}

		event := socketmode.Event{}
		event.Type = socketmode.EventTypeInteractive
		event.Data = callback

		bot.handleSocketModeEvent(event)
	})

	t.Run("handle unauthorized interaction", func(t *testing.T) {
		callback := slack.InteractionCallback{
			User: slack.User{
				ID: "unknown",
			},
		}

		success := bot.handleInteraction(callback)
		assert.False(t, success)
	})

	t.Run("load file content without file", func(t *testing.T) {
		event := &slackevents.MessageEvent{}
		actual := bot.loadFileContent(event)
		assert.Equal(t, "", actual)
	})
}

func TestReplaceClickedButton(t *testing.T) {
	messageEvent := &slack.MessageEvent{}

	action := slack.NewActionBlock("", client.GetInteractionButton("reply", "my text", "replay YEP", slack.StylePrimary))
	button := action.Elements.ElementSet[0].(*slack.ButtonBlockElement)
	actionID := button.Value
	assert.Equal(t, "replay YEP", actionID)

	messageEvent.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			action,
		},
	}

	actual := replaceClickedButton((*slack.Message)(messageEvent), actionID, " (worked)")
	jsonString, err := json.Marshal(actual)

	expected := `{"replace_original":false,"delete_original":false,"metadata":{"event_type":"","event_payload":null},"blocks":[{"type":"actions","elements":[{"type":"button","text":{"type":"plain_text","text":"my text (worked)","emoji":true},"action_id":"reply","style":"danger"}]}]}`

	require.NoError(t, err)
	assert.Equal(t, expected, string(jsonString))
}
