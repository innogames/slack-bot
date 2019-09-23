package command

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/nlopes/slack"
	"log"
	"net/http"
	"time"
)

func NewWeatherCommand(slackClient client.SlackClient, config config.OpenWeather) bot.Command {
	return WeatherCommand{slackClient, config}
}

type City struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type Weather struct {
	Id          int    `json:"id"`
	Main        string `json:"main"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Wind struct {
	Speed float64 `json:"speed"`
	Deg   float64 `json:"deg"`
}

type Clouds struct {
	All int `json:"all"`
}

type Rain struct {
	Threehr int `json:"3h"`
}

type Main struct {
	Temp     float64 `json:"temp"`
	Pressure int     `json:"pressure"`
	Humidity int     `json:"humidity"`
	Temp_min float64 `json:"temp_min"`
	Temp_max float64 `json:"temp_max"`
}

type Sys struct {
	Country string `json:"country"`
	Sunrise int    `json:"sunrise"`
	Sunset  int    `json:"sunset"`
}

type CurrentWeatherResponse struct {
	Coord   Coord     `json:"coord"`
	Weather []Weather `json:"weather"`
	Main    Main      `json:"main"`
	Wind    Wind      `json:"wind"`
	Rain    Rain      `json:"rain"`
	Clouds  Clouds    `json:"clouds"`
	Sys     Sys       `json:"sys"`
	Dt      int       `json:"dt"`
	Id      int       `json:"id"`
	Name    string    `json:"name"`
}

type ForecastResponse struct {
	City    `json:"city"`
	Coord   `json:"coord"`
	Country string `json:"country"`
	List    []struct {
		Dt      int `json:"dt"`
		Main    `json:"main"`
		Weather `json:"weather"`
		Clouds  `json:"clouds"`
		Wind    `json:"wind"`
	} `json:"list"`
}

type WeatherCommand struct {
	slackClient client.SlackClient
	config      config.OpenWeather
}

func (c WeatherCommand) GetMatcher() matcher.Matcher {
	return matcher.NewRegexpMatcher("weather", c.GetWeather)
}

func (c WeatherCommand) GetWeather(match matcher.Result, event slack.MessageEvent) {
	url := fmt.Sprintf("%s?q=%s&units=%s&appid=%s", c.config.Url, c.config.Location, c.config.Units, c.config.Apikey)
	response, err := http.Get(url)
	if err != nil {
		c.slackClient.Reply(event, "Api call returned an err: "+err.Error())
	}

	var record CurrentWeatherResponse
	err = json.NewDecoder(response.Body).Decode(&record)
	if err != nil {
		log.Println(err)
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

func (c WeatherCommand) GetHelp() []bot.Help {
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
