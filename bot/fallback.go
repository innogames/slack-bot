package bot

import (
	"fmt"
	"math"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const minDistance = 4

// try to find the best matching commands based on command name and examples
func (b *Bot) sendFallbackMessage(message msg.Message) {
	bestMatching := getBestMatchingHelp(b, message.Text)

	if bestMatching.Command == "" {
		b.slackClient.AddReaction("❓", message)

		blocks := []slack.Block{
			client.GetTextBlock(
				"Oops! Command `" + message.Text + "` not found...try `help`.",
			),
			slack.NewActionBlock(
				"",
				client.GetInteractionButton("Help!", "help"),
			),
		}
		b.slackClient.SendBlockMessage(message, blocks)

		return
	}

	b.slackClient.SendMessage(message, "Command `"+message.Text+"` not found...do you mean *"+bestMatching.Command+"* command?")

	client.HandleMessage(
		message.WithText(fmt.Sprintf("help %s", bestMatching.Command)),
	)
}

// find the best matching command bases on the given strings...using levenstein to fetch the best one
func getBestMatchingHelp(b *Bot, eventText string) Help {
	distance := math.MaxInt32
	var bestMatching Help

	eventText = strings.ToLower(eventText)

	for _, commandHelp := range b.commands.GetHelp() {
		for _, token := range commandHelp.GetKeywords() {
			currentDistance := levenshtein.DistanceForStrings(
				[]rune(strings.ToLower(token)),
				[]rune(eventText),
				levenshtein.DefaultOptions,
			)

			if currentDistance <= minDistance && currentDistance < distance {
				bestMatching = commandHelp
				distance = currentDistance
			}
		}
	}

	return bestMatching
}
