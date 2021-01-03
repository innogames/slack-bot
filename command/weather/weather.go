package weather

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/client"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

const defaultAPIURL = "https://api.openweathermap.org/data/2.5/weather"

// NewWeatherCommand is using OpenWeatherMap to display current weather and the forecast
func NewWeatherCommand(base bot.BaseCommand, cfg config.OpenWeather) bot.Command {
	if cfg.URL == "" {
		cfg.URL = defaultAPIURL
	}

	return &command{base, cfg}
}

type command struct {
	bot.BaseCommand
	cfg config.OpenWeather
}

func (c *command) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("weather", c.GetWeather),
		matcher.NewRegexpMatcher(`weather in (?P<location>\w\s+)`, c.GetWeather),
	)
}

func (c *command) GetWeather(match matcher.Result, message msg.Message) {
	location := match.GetString("location")
	if location == "" {
		location = c.cfg.Location
	}

	apiURL := fmt.Sprintf(
		"%s?q=%s&units=%s&appid=%s",
		c.cfg.URL,
		url.QueryEscape(location),
		c.cfg.Units,
		c.cfg.Apikey,
	)

	response, err := client.HTTPClient.Get(apiURL)
	if err != nil {
		c.ReplyError(message, errors.Wrap(err, "Api call returned an err"))
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 300 {
		c.SendMessage(message, fmt.Sprintf("Api call returned an err: %d", response.StatusCode))
		return
	}

	var record CurrentWeatherResponse
	err = json.NewDecoder(response.Body).Decode(&record)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	var fields = [][]string{
		{
			fmt.Sprintf("*:thermometer: TEMPERATURE:*  *Current:* %2.fC°", record.Main.Temp),
			fmt.Sprintf(":arrow_down_small: *Min:* %2.fC°\t\t:arrow_up_small: *Max:* %2.fC°", record.Main.TempMin, record.Main.TempMax),
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
			fmt.Sprintf("*:city_sunrise: Sunrise:* %s :clock7:", timestampToTime(record.Sys.Sunrise)),
			fmt.Sprintf("*:night_with_stars: Sunset:* %s :clock11: ", timestampToTime(record.Sys.Sunset)),
		},
	}

	headerText := slack.NewTextBlockObject(
		"mrkdwn",
		fmt.Sprintf(
			"*Weather information for: %s :flag-%s:* %s",
			record.Name,
			record.Sys.Country,
			getIcon(record.Weather[0].ID),
		),
		false,
		false,
	)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	sections := make([]slack.Block, 0, len(fields)+1)
	sections = append(sections, headerSection)

	for _, element := range fields {
		var textBlocks []*slack.TextBlockObject
		textBlocks = append(
			textBlocks,
			slack.NewTextBlockObject("mrkdwn", element[0], false, false),
			slack.NewTextBlockObject("mrkdwn", element[1], false, false),
		)

		sections = append(sections, slack.NewSectionBlock(nil, textBlocks, nil))
	}

	c.SendMessage(message, "", slack.MsgOptionBlocks(sections...))
}

func timestampToTime(timestamp int) string {
	return time.Unix(int64(timestamp), 0).Format("15:04")
}

func (c *command) IsEnabled() bool {
	return c.cfg.Apikey != ""
}

func (c *command) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "weather",
			Description: "returns the current weather information",
			Examples: []string{
				"weather",
				"weather in Berlin",
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
