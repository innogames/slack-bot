package command

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/admin"
	"github.com/innogames/slack-bot/command/cron"
	"github.com/innogames/slack-bot/command/custom"
	"github.com/innogames/slack-bot/command/games"
	"github.com/innogames/slack-bot/command/jenkins"
	"github.com/innogames/slack-bot/command/jira"
	"github.com/innogames/slack-bot/command/pullrequest"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/innogames/slack-bot/command/variables"
	"github.com/innogames/slack-bot/command/weather"
)

// GetCommands returns the list of default command which are available
func GetCommands(slackClient client.SlackClient, cfg config.Config) *bot.Commands {
	commands := &bot.Commands{}
	commands.AddCommand(
		// needs to be the first commands to store all executed commands
		NewRetryCommand(slackClient),

		NewCommands(slackClient, cfg.Commands),
		NewReplyCommand(slackClient),
		NewAddLinkCommand(slackClient),
		NewAddButtonCommand(slackClient, cfg.Server),
		NewReactionCommand(slackClient),
		NewSendMessageCommand(slackClient),
		NewDelayCommand(slackClient),
		NewRandomCommand(slackClient),
		NewHelpCommand(slackClient, commands),

		admin.NewBotLogCommand(slackClient, cfg),
		admin.NewStatsCommand(slackClient, cfg),

		weather.NewWeatherCommand(slackClient, cfg.OpenWeather),

		cron.NewCronCommand(slackClient, cfg.Crons),

		queue.NewQueueCommand(slackClient),
		queue.NewListCommand(slackClient),

		custom.GetCommand(slackClient),
		variables.GetCommand(slackClient),
	)

	// games
	commands.Merge(games.GetCommands(slackClient))

	// jira
	commands.Merge(jira.GetCommands(&cfg.Jira, slackClient))

	// jenkins
	commands.Merge(jenkins.GetCommands(cfg.Jenkins, slackClient))

	// pull-request
	commands.Merge(pullrequest.GetCommands(slackClient, cfg))

	return commands
}
