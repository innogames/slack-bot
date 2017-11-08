package jenkins

import (
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/client/jenkins"
	"github.com/nlopes/slack"
)

type nodesCommand struct {
	jenkins     jenkins.Client
	slackClient client.SlackClient
}

func NewNodesCommand(jenkins jenkins.Client, slackClient client.SlackClient) bot.Command {
	return &nodesCommand{jenkins, slackClient}
}

func (c *nodesCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("jenkins nodes", c.Run)
}

func (c *nodesCommand) IsEnabled() bool {
	return c.jenkins != nil
}

func (c *nodesCommand) Run(match matcher.Result, event slack.MessageEvent) {
	nodes, err := c.jenkins.GetAllNodes()
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	text := "*Nodes*\n"
	for _, node := range nodes {
		offline := node.Raw.Offline
		text += fmt.Sprintf("- *%s* - online: *%t* - executors: %d\n", node.GetName(), !offline, len(node.Raw.Executors))
	}

	c.slackClient.Reply(event, text)
}

func (c *nodesCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			"jenkins nodes",
			"Prints a list of all jenkins nodes",
			[]string{
				"jenkins nodes",
			},
		},
	}
}
