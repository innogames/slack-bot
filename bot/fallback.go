package bot

import (
	"fmt"
	"math"
	"strings"

	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
	"github.com/texttheater/golang-levenshtein/levenshtein"
)

const minDistance = 4

// try to find the best matching commands based on command name and examples
func (b bot) sendFallbackMessage(event slack.MessageEvent) {
	bestMatching := getBestMatchingHelp(b, event.Text)

	if bestMatching.Command == "" {
		b.slackClient.SendMessage(event, "Oops! Command `"+event.Text+"` not found...try `help`.")
		return
	}

	b.slackClient.SendMessage(event, "Command `"+event.Text+"` not found...do you mean *"+bestMatching.Command+"* command?")

	event.Text = fmt.Sprintf("help %s", bestMatching.Command)
	client.InternalMessages <- event
}

func getBestMatchingHelp(b bot, eventText string) Help {
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
