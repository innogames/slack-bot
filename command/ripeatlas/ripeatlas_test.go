package ripeatlas

import (
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

func spawnRIPEAtlasServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	authenticate := func(res http.ResponseWriter, req *http.Request) bool {
		assert.NotNil(t, req.Header.Get("Authorization"))

		authHeader := req.Header.Get("Authorization")
		if authHeader != "Key apikey" {
			res.WriteHeader(http.StatusForbidden)
			res.Write([]byte(`{"error":{"detail":"The provided API key does not exist","status":403,"title":"Forbidden","code":104}}`))
			return false
		}
		return true
	}

	// test connection
	mux.HandleFunc("/", func(res http.ResponseWriter, r *http.Request) {
		// mock stream response
		if r.URL.RawQuery == "streamType=result&msm=58913886" {
			res.Write([]byte(`["atlas_subscribed",{"streamType":"result","msm":1001}]` + "\n"))
			res.Write([]byte(`["atlas_result",{"msm_id": 58913886,"timestamp":1234567890,"result":[{"hop":1,"result":[{"from":"whatever1.host","size":100},{"from":"whatever2.host","size":100},{"from":"whatever3.host","size":100}]}]}]` + "\n"))
			return
		}

		res.Write([]byte(`ok`))
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
	slackClient := mocks.NewSlackClient(t)
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
		ts := spawnRIPEAtlasServer(t)
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
		ts := spawnRIPEAtlasServer(t)
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

	t.Run("RIPE Atlas Traceroute Destination Parsing", func(t *testing.T) {
		assert.Equal(t, parseDestination("8.8.8.8"), 4)
		assert.Equal(t, parseDestination("2001:4860:4860::8844"), 6)
		assert.Equal(t, parseDestination("example.com"), 6)
	})

	t.Run("RIPE Atlas Traceroute API wrong key", func(t *testing.T) {
		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t)
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
		// Set a proper timezone, otherwise the test fails on GitHub Actions
		time.Local, _ = time.LoadLocation("Europe/Berlin")

		// mock RIPE Atlas API
		ts := spawnRIPEAtlasServer(t)
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
		expectedResult := "```\n" +
			"Start: 2009-02-14 00:31:30 +0100 CET\n" +
			"HOST:                                          Loss%  RTT\n" +
			" 1 .  whatever1.host                              0%    0.000   0.000   0.000\n" +
			"```\n"
		mocks.AssertSlackMessage(slackClient, message, expectedResult, mock.Anything)

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})
}
