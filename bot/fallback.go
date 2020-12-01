package bot

import (
	"fmt"
	"github.com/innogames/slack-bot/bot/msg"
	"math"
	"strings"

	"github.com/innogames/slack-bot/client"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const minDistance = 4

// try to find the best matching commands based on command name and examples
func (b *Bot) sendFallbackMessage(message msg.Message) {
	bestMatching := getBestMatchingHelp(b, message.Text)

	if bestMatching.Command == "" {
		b.slackClient.SendMessage(message, "Oops! Command `"+message.Text+"` not found...try `help`.")
		return
	}

	b.slackClient.SendMessage(message, "Command `"+message.Text+"` not found...do you mean *"+bestMatching.Command+"* command?")

	client.InternalMessages <- message.WithText(fmt.Sprintf("help %s", bestMatching.Command))
}

// find the best matching command bases on the given strings...using levenstein to fetch the best one
func getBestMatchingHelp(b *Bot, eventText string) Help {
	var distance = math.MaxInt32
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
