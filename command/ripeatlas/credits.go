package ripeatlas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
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

func (c *creditsCommand) credits(_ matcher.Result, message msg.Message) {
	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	url := c.cfg.APIURL + "/credits"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("request creation returned an err: %w", err))
		log.Errorf("request creation returned an err: %s", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Key "+c.cfg.APIKey)

	response, err := client.GetHTTPClient().Do(req)
	if err != nil {
		c.ReplyError(message, fmt.Errorf("API call returned an err: %w", err))
		log.Errorf("API call returned an err: %s", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		c.ReplyError(message, fmt.Errorf("API call returned an err: %d", response.StatusCode))
		log.Errorf("API call returned an err: %d", response.StatusCode)
		return
	}

	var result CreditsResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		c.ReplyError(message, err)
		log.Errorf("%s", err)
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
