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
	base := bot.BaseCommand{SlackClient: slackClient}

	commands := &bot.Commands{}
	commands.AddCommand(
		// needs to be the first commands to store all executed commands
		NewRetryCommand(base, &cfg),

		NewCommands(base, cfg.Commands),
		NewReplyCommand(base),
		NewAddLinkCommand(base),
		NewAddButtonCommand(base, cfg.Slack),
		NewReactionCommand(base),
		NewSendMessageCommand(base),
		NewDelayCommand(base),
		NewRandomCommand(base),
		NewHelpCommand(base, commands),

		admin.NewBotLogCommand(base, cfg),
		admin.NewStatsCommand(base, cfg),

		weather.NewWeatherCommand(base, cfg.OpenWeather),

		cron.NewCronCommand(base, cfg.Crons),

		queue.NewQueueCommand(base),
		queue.NewListCommand(base),

		custom.GetCommand(base),
		variables.GetCommand(base),
	)

	// games
	commands.Merge(games.GetCommands(base))

	// jira
	commands.Merge(jira.GetCommands(&cfg.Jira, slackClient))

	// jenkins
	commands.Merge(jenkins.GetCommands(cfg.Jenkins, base))

	// pull-request
	commands.Merge(pullrequest.GetCommands(base, &cfg))

	return commands
}
