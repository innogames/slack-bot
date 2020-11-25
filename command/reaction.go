package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

// NewReactionCommand simply adds a reaction to the current message...used in "commands" and other internal commands
func NewReactionCommand(slackClient client.SlackClient) bot.Command {
	return &reactionCommand{slackClient}
}

type reactionCommand struct {
	slackClient client.SlackClient
}

func (c *reactionCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`add reaction :(?P<reaction>.*):`, c.Add),
		matcher.NewRegexpMatcher(`remove reaction :(?P<reaction>.*):`, c.Remove),
	)
}

func (c *reactionCommand) Add(match matcher.Result, event slack.MessageEvent) {
	reaction := match.GetString("reaction")
	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
	c.slackClient.AddReaction(reaction, msgRef)
}

func (c *reactionCommand) Remove(match matcher.Result, event slack.MessageEvent) {
	reaction := match.GetString("reaction")
	msgRef := slack.NewRefToMessage(event.Channel, event.Timestamp)
	c.slackClient.RemoveReaction(reaction, msgRef)
}

func (c *reactionCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "add reaction",
			Description: "add a reaction on a message",
			Examples: []string{
				"add reaction :check_mark:",
			},
		},
		{
			Command:     "remove reaction",
			Description: "remove a reaction on a message",
			Examples: []string{
				"remove reaction :check_mark:",
			},
		},
	}
}
