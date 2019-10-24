package weather

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

const defaultApiUrl = "https://api.openweathermap.org/data/2.5/weather"

// NewWeatherCommand is using OpenWeatherMap to display current weather and the forecast
func NewWeatherCommand(slackClient client.SlackClient, config config.OpenWeather) bot.Command {
	if config.Url == "" {
		config.Url = defaultApiUrl
	}

	return command{slackClient, config, http.Client{}}
}

type command struct {
	slackClient client.SlackClient
	config      config.OpenWeather
	client      http.Client
}

func (c command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("weather", c.GetWeather),
		matcher.NewRegexpMatcher("weather in (?P<location>\\w\\s+)", c.GetWeather),
	)
}

func (c command) GetWeather(match matcher.Result, event slack.MessageEvent) {
	location := match.GetString("location")
	if location == "" {
		location = c.config.Location
	}

	url := fmt.Sprintf(
		"%s?q=%s&units=%s&appid=%s",
		c.config.Url,
		url.QueryEscape(location),
		c.config.Units,
		c.config.Apikey,
	)

	response, err := c.client.Get(url)
	if err != nil {
		c.slackClient.ReplyError(event, errors.Wrap(err, "Api call returned an err"))
		return
	}

	if response.StatusCode >= 300 {
		c.slackClient.Reply(event, fmt.Sprintf("Api call returned an err: %d", response.StatusCode))
		return
	}

	var record CurrentWeatherResponse
	err = json.NewDecoder(response.Body).Decode(&record)
	if err != nil {
		c.slackClient.ReplyError(event, err)
		return
	}

	var fields = [][]string{
		{
			fmt.Sprintf("*:thermometer: TEMPERATURE:*  *Current:* %2.fC°", record.Main.Temp),
			fmt.Sprintf(":arrow_down_small: *Min:* %2.fC°\t\t:arrow_up_small: *Max:* %2.fC°", record.Main.Temp_min, record.Main.Temp_max),
		},
		{
			fmt.Sprintf("*:wind_blowing_face: WIND:*  *Speed:* %2.fm/s", record.Wind.Speed),
			fmt.Sprintf(":arrow_right: *Direction:* %2.f°", record.Wind.Deg),
		},
		{
			fmt.Sprintf("*:sweat_drops: RAIN:*  *Humidity:* %d%%", record.Main.Humidity),
			fmt.Sprintf("*:cloud: Clouds:* %d%%", record.Clouds.All),
		},
		{
			fmt.Sprintf("*:city_sunrise: Sunrise:* %s :clock7:", time.Unix(int64(record.Sys.Sunrise), 0).Format("15:15")),
			fmt.Sprintf("*:night_with_stars: Sunset:* %s :clock11: ", time.Unix(int64(record.Sys.Sunset), 0).Format("15:15")),
		},
	}

	headerText := slack.NewTextBlockObject(
		"mrkdwn",
		fmt.Sprintf(
			"*Weather information for: %s :flag-%s:* %s",
			record.Name,
			record.Sys.Country,
			getIcon(record.Weather[0].Id),
		),
		false,
		false,
	)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	var sections []slack.Block
	sections = append(sections, headerSection)
	for _, element := range fields {
		var textBlocks []*slack.TextBlockObject
		textBlocks = append(textBlocks, slack.NewTextBlockObject("mrkdwn", element[0], false, false))
		textBlocks = append(textBlocks, slack.NewTextBlockObject("mrkdwn", element[1], false, false))

		sections = append(sections, slack.NewSectionBlock(nil, textBlocks, nil))
	}

	c.slackClient.SendMessage(event, "", slack.MsgOptionBlocks(sections...))
}

func (c command) IsEnabled() bool {
	return c.config.Apikey != ""
}

func (c command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "weather",
			Description: "returns the current weather information",
			Examples: []string{
				"weather",
			},
		},
	}
}

func getIcon(weatherCode int) string {
	switch {
	case weatherCode > 800:
		return ":cloud:"
	case weatherCode == 800:
		return ":sunny:"
	case weatherCode >= 700:
		return ":fog:"
	case weatherCode >= 600:
		return ":snow_cloud:"
	case weatherCode >= 300:
		return ":rain_cloud:"
	case weatherCode >= 200:
		return ":thunder_cloud_and_rain:"
	default:
		return ":sunny:"
	}
}
