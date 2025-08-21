package command

import (
    "context"
    "errors"
    "fmt"
    "net/http"

    "github.com/innogames/slack-bot/v2/bot"
    "github.com/innogames/slack-bot/v2/bot/matcher"
    "github.com/innogames/slack-bot/v2/bot/msg"
    "github.com/innogames/slack-bot/v2/client"
)

func NewRequestCommand(base bot.BaseCommand) bot.Command {
	return &requestCommand{base}
}

type requestCommand struct {
	bot.BaseCommand
}

func (c requestCommand) GetMatcher() matcher.Matcher {
	return matcher.NewOptionMatcher(
		"request",
		[]string{"method", "url"},
		c.doRequest,
		c.SlackClient,
	)
}

func (c requestCommand) doRequest(match matcher.Result, message msg.Message) {
	method := match.GetString("method")
	if method == "" {
		method = "GET"
	}
    url := match.GetString("url")
	if url == "" {
		c.ReplyError(message, errors.New("please provide a valid url"))
		return
	}

    request, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
        c.ReplyError(message, fmt.Errorf("invalid request: %w", err))
		return
	}

	httpClient := client.GetHTTPClient()
    response, err := httpClient.Do(request)
	if err != nil {
		c.AddReaction("❌", message)
        c.ReplyError(message, fmt.Errorf("request failed: %w", err))
		return
	}
	defer response.Body.Close()

	// check success status
	if response.StatusCode >= 400 {
		c.AddReaction("❌", message)
		c.ReplyError(message, errors.New("request failed with status "+response.Status))
		return
	}

	c.AddReaction("white_check_mark", message)
}

func (c requestCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "request --url=<url> [--method=<method>]",
			Description: "send a request to the given url",
			Examples: []string{
				"request GET https://example.com",
				"request POST https://jenkins.exmaple.com/webhook?auth=1",
			},
		},
	}
}
