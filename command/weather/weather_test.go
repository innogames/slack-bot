package weather

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWeather(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	// set default timezone
	time.Local, _ = time.LoadLocation("Europe/Berlin")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		file, _ := os.ReadFile("./dump_current_weather.json")
		w.Write(file)
	}))
	defer ts.Close()

	cfg := config.OpenWeather{}
	cfg.Location = "Hamburg"
	cfg.Apikey = "12345"
	cfg.URL = ts.URL

	command := bot.Commands{}
	command.AddCommand(NewWeatherCommand(base, cfg))

	t.Run("Send invalid command", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "I hate the current weather..."

		actual := command.Run(message)
		assert.False(t, actual)
	})

	t.Run("Fetch default weather", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "weather"

		expected := `[` +
			`{"type":"section","text":` +
			`{"type":"mrkdwn","text":"*Weather information for: Hamburg :flag-DE:* :cloud:"}},` +
			`{"type":"section","fields":[{"type":"mrkdwn","text":"*:thermometer: TEMPERATURE:*  *Current:* 288C°"},` +
			`{"type":"mrkdwn","text":":arrow_down_small: *Min:* 286C°\t\t:arrow_up_small: *Max:* 290C°"}]},` +
			`{"type":"section","fields":[{"type":"mrkdwn","text":"*:wind_blowing_face: WIND:*  *Speed:*  4m/s"},` +
			`{"type":"mrkdwn","text":":arrow_right: *Direction:* 210°"}]},` +
			`{"type":"section","fields":[{"type":"mrkdwn","text":"*:sweat_drops: RAIN:*  *Humidity:* 87%"},` +
			`{"type":"mrkdwn","text":"*:cloud: Clouds:* 75%"}]},` +
			`{"type":"section","fields":[{"type":"mrkdwn","text":"*:city_sunrise: Sunrise:* 07:57 :clock7:"},` +
			`{"type":"mrkdwn","text":"*:night_with_stars: Sunset:* 18:11 :clock11: "}]}]`

		mocks.AssertSlackBlocks(t, slackClient, message, expected)

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test help", func(t *testing.T) {
		help := command.GetHelp()
		assert.Len(t, help, 2)
	})

	t.Run("Fetch with invalid API-Key", func(t *testing.T) {
		message := msg.Message{}
		message.Text = "weather"

		cfg := config.OpenWeather{}
		cfg.Location = "Hamburg"
		cfg.Apikey = "12345"

		mocks.AssertSlackMessage(slackClient, message, "Api call returned an err: 401")

		command := bot.Commands{}
		command.AddCommand(NewWeatherCommand(base, cfg))

		actual := command.Run(message)
		assert.True(t, actual)
	})

	t.Run("Test weather icon", func(t *testing.T) {
		// tests for each possible icon code that we have a valid reaction/emoji
		for code := range 1001 {
			icon := getIcon(code)

			reaction := util.Reaction(icon)
			assert.NotEqual(t, "?", reaction)
		}
	})
}
