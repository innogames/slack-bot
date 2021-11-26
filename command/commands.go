package command

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command/admin"
	"github.com/innogames/slack-bot/v2/command/clouds/aws"
	"github.com/innogames/slack-bot/v2/command/cron"
	"github.com/innogames/slack-bot/v2/command/custom"
	"github.com/innogames/slack-bot/v2/command/games"
	"github.com/innogames/slack-bot/v2/command/jenkins"
	"github.com/innogames/slack-bot/v2/command/jira"
	"github.com/innogames/slack-bot/v2/command/pullrequest"
	"github.com/innogames/slack-bot/v2/command/queue"
	"github.com/innogames/slack-bot/v2/command/variables"
	"github.com/innogames/slack-bot/v2/command/weather"
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
		NewAddButtonCommand(base),
		NewReactionCommand(base),
		NewSendMessageCommand(base),
		NewDelayCommand(base),
		NewRandomCommand(base),
		NewHelpCommand(base, commands),

		weather.NewWeatherCommand(base, cfg.OpenWeather),

		cron.NewCronCommand(base, cfg.Crons),

		queue.NewQueueCommand(base),
		queue.NewListCommand(base),

		custom.GetCommand(base),
		variables.GetCommand(base),
	)

	commands.Merge(admin.GetCommands(base, &cfg))

	// games
	commands.Merge(games.GetCommands(base))

	// jira
	commands.Merge(jira.GetCommands(&cfg.Jira, slackClient))

	// jenkins
	commands.Merge(jenkins.GetCommands(cfg.Jenkins, base))

	// pull-request
	commands.Merge(pullrequest.GetCommands(base, &cfg))

	// aws
	commands.Merge(aws.GetCommands(cfg.Aws, base))

	return commands
}

// helperCategory is used to group all helper commands which are usable in crons/commands etc
var helperCategory = bot.Category{
	Name:        "Helper",
	Description: "usable e.g. in 'commands' or 'crons'",
	HelpURL:     "https://github.com/innogames/slack-bot#custom-command",
}
