package ripeatlas

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func spawnRIPEAtlasServer(t *testing.T, apikey string) *httptest.Server {
	mux := http.NewServeMux()

	authenticate := func(res http.ResponseWriter, req *http.Request) bool {
		assert.NotNil(t, req.Header.Get("Authorization"))

		authHeader := req.Header.Get("Authorization")
		if authHeader != fmt.Sprintf("Key %s", apikey) {
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte(`{"error":{"detail":"The provided API key does not exist","status":403,"title":"Forbidden","code":104}}`))
			return false
		}
		return true
	}

	// test connection
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	mux.HandleFunc("/credits", func(res http.ResponseWriter, req *http.Request) {
		// Ensure body is empty
		givenInputJSON, _ := io.ReadAll(req.Body)
		assert.Empty(t, string(givenInputJSON))

		// Check for authentication
		if !authenticate(res, req) {
			return
		}

		res.WriteHeader(http.StatusOK)
		res.Write([]byte(`{		"current_balance": 998060,
									"credit_checked": true,
									"max_daily_credits": 1000000,
									"estimated_daily_income": 0,
									"estimated_daily_expenditure": 0,
									"estimated_daily_balance": 0,
									"calculation_time": "2023-08-23T15:15:49.274480",
									"estimated_runout_seconds": null,
									"past_day_measurement_results": 3,
									"past_day_credits_spent": 180,
									"last_date_debited": "2023-08-22T00:18:47.516307",
									"last_date_credited": "2022-03-10T22:22:28.728145",
									"income_items": "https://atlas.ripe.net/api/v2/credits/income-items/",
									"expense_items": "https://atlas.ripe.net/api/v2/credits/expense-items/",
									"transactions": "https://atlas.ripe.net/api/v2/credits/transactions/"}`))
	})

	mux.HandleFunc("/measurements", func(res http.ResponseWriter, req *http.Request) {
		// Check for authentication
		if !authenticate(res, req) {
			return
		}

		expectedJSON := `{"definitions":[{"target":"8.8.8.8","af":4,"description":"Slackbot measurement to 8.8.8.8","protocol":"ICMP","packets":3,"type":"traceroute","is_public":true}],"probes":[{"type":"area","value":"WW","requested":1}],"is_oneoff":true}`
		responseJSON := `{"measurements":[58913886]}`

		givenInputJSON, _ := io.ReadAll(req.Body)
		assert.Equal(t, expectedJSON, string(givenInputJSON))

		res.WriteHeader(http.StatusOK)
		res.Write([]byte(responseJSON))
	})

	return httptest.NewServer(mux)
}

func TestRipeAtlas(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("RIPE Atlas is not active", func(t *testing.T) {
		cfg := &config.Config{}
		commands := GetCommands(base, cfg)
		assert.Equal(t, 0, commands.Count())
	})

	t.Run("RIPE Atlas is active", func(t *testing.T) {
		ripeAtlasCfg := defaultConfig
		ripeAtlasCfg.APIKey = "apikey"

		cfg := &config.Config{}
		cfg.Set("ripeatlas", ripeAtlasCfg)
		commands := GetCommands(base, cfg)
		assert.Equal(t, 2, commands.Count())

		help := commands.GetHelp()
		assert.Equal(t, 2, len(help))
	})

	t.Run("RIPE Atlas Credits API wrong key", func(t *testing.T) {
		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t, "apikey")
		defer ts.Close()

		ripeAtlasCfg := defaultConfig
		ripeAtlasCfg.APIKey = "nope"
		ripeAtlasCfg.APIURL = ts.URL

		cfg := &config.Config{}
		cfg.Set("ripeatlas", ripeAtlasCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "credits"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertError(slackClient, message, "API call returned an err: 403")

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})

	t.Run("RIPE Atlas Credits API works", func(t *testing.T) {
		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t, "apikey")
		defer ts.Close()

		ripeAtlasCfg := defaultConfig
		ripeAtlasCfg.APIKey = "apikey"
		ripeAtlasCfg.APIURL = ts.URL

		cfg := &config.Config{}
		cfg.Set("ripeatlas", ripeAtlasCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "credits"

		mocks.AssertReaction(slackClient, ":coffee:", message)
		mocks.AssertRemoveReaction(slackClient, ":coffee:", message)
		mocks.AssertSlackMessage(slackClient, message, "Total credits remaining: 998060")

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})

	t.Run("RIPE Atlas Traceroute API wrong key", func(t *testing.T) {
		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t, "apikey")
		defer ts.Close()

		ripeAtlasCfg := defaultConfig
		ripeAtlasCfg.APIKey = "nope"
		ripeAtlasCfg.APIURL = ts.URL
		ripeAtlasCfg.StreamURL = ts.URL

		cfg := &config.Config{}
		cfg.Set("ripeatlas", ripeAtlasCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "traceroute 8.8.8.8"

		mocks.AssertReaction(slackClient, ":stopwatch:", message)
		mocks.AssertRemoveReaction(slackClient, ":stopwatch:", message)
		mocks.AssertError(slackClient, message, "API call returned an err: 403")

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})

	t.Run("RIPE Atlas Traceroute API works", func(t *testing.T) {
		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t, "apikey")
		defer ts.Close()

		ripeAtlasCfg := defaultConfig
		ripeAtlasCfg.APIKey = "apikey"
		ripeAtlasCfg.APIURL = ts.URL
		ripeAtlasCfg.StreamURL = ts.URL

		cfg := &config.Config{}
		cfg.Set("ripeatlas", ripeAtlasCfg)
		commands := GetCommands(base, cfg)

		message := msg.Message{}
		message.Text = "traceroute 8.8.8.8"

		mocks.AssertReaction(slackClient, ":stopwatch:", message)
		mocks.AssertRemoveReaction(slackClient, ":stopwatch:", message)
		mocks.AssertSlackMessage(slackClient, message, "Measurement created: https://atlas.ripe.net/measurements/58913886\n", mock.Anything)
		// TODO: Check for the streaming message here as well. Currently it's broken due to a bug on the mock.

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})
}
