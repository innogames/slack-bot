package weather

import (
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/mocks"
	"github.com/nlopes/slack"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestWeather(t *testing.T) {
	slackClient := &mocks.SlackClient{}

	// set default timezone
	time.Local, _ = time.LoadLocation("Europe/Berlin")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("./dump_current_weather.json")
		w.Write(file)
	}))
	defer ts.Close()

	cfg := config.OpenWeather{}
	cfg.Location = "Hamburg"
	cfg.Apikey = "12345"
	cfg.Url = ts.URL

	command := bot.Commands{}
	command.AddCommand(NewWeatherCommand(slackClient, cfg))

	t.Run("Send invalid command", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "I hate the current weather..."

		actual := command.Run(event)
		assert.Equal(t, false, actual)
	})

	t.Run("Fetch default weather", func(t *testing.T) {
		event := slack.MessageEvent{}
		event.Text = "weather"

		expected := url.Values{}
		expected.Add("blocks", "["+
			"{\"type\":\"section\",\"text\":"+
			"{\"type\":\"mrkdwn\",\"text\":\"*Weather information for: Hamburg :flag-DE:* :cloud:\"}},"+
			"{\"type\":\"section\",\"fields\":[{\"type\":\"mrkdwn\",\"text\":\"*:thermometer: TEMPERATURE:*  *Current:* 288C°\"},"+
			"{\"type\":\"mrkdwn\",\"text\":\":arrow_down_small: *Min:* 286C°\\t\\t:arrow_up_small: *Max:* 290C°\"}]},"+
			"{\"type\":\"section\",\"fields\":[{\"type\":\"mrkdwn\",\"text\":\"*:wind_blowing_face: WIND:*  *Speed:*  4m/s\"},"+
			"{\"type\":\"mrkdwn\",\"text\":\":arrow_right: *Direction:* 210°\"}]},"+
			"{\"type\":\"section\",\"fields\":[{\"type\":\"mrkdwn\",\"text\":\"*:sweat_drops: RAIN:*  *Humidity:* 87%\"},"+
			"{\"type\":\"mrkdwn\",\"text\":\"*:cloud: Clouds:* 75%\"}]},"+
			"{\"type\":\"section\",\"fields\":[{\"type\":\"mrkdwn\",\"text\":\"*:city_sunrise: Sunrise:* 07:57 :clock7:\"},"+
			"{\"type\":\"mrkdwn\",\"text\":\"*:night_with_stars: Sunset:* 18:11 :clock11: \"}]}]",
		)

		mocks.AssertSlackJson(t, slackClient, event, expected)

		actual := command.Run(event)
		assert.Equal(t, true, actual)
	})
}
