package ripeatlas

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/pkg/errors"
	"net/http"
)

type creditsCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *creditsCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("credits", c.credits),
	)
}

func (c *creditsCommand) credits(match matcher.Result, message msg.Message) {

	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	url := fmt.Sprintf("%s/credits", c.cfg.APIURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.ReplyError(message, errors.Wrap(err, "Request creation returned an err"))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.cfg.APIKey)

	response, err := client.GetHTTPClient().Do(req)
	if err != nil {
		c.ReplyError(message, errors.Wrap(err, "Api call returned an err"))
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		c.SendMessage(message, fmt.Sprintf("Api call returned an err: %d", response.StatusCode))
		return
	}

	var result CreditsResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	text := fmt.Sprintf("Total credits remaining: %d", result.CurrentBalance)

	c.SendMessage(message, text)
}

func (c *creditsCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "credits",
			Description: "Query how many credits are available for this API key",
			Category:    category,
		},
	}
}
