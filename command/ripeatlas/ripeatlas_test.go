package ripeatlas

import (
	"fmt"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

type testRequest struct {
	inputJSON    string
	responseJSON string
	responseCode int
}

func startTestServer(t *testing.T, requests []testRequest) *httptest.Server {
	t.Helper()

	idx := 0

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		expected := requests[idx]
		idx++

		givenInputJSON, _ := io.ReadAll(req.Body)

		fmt.Println(givenInputJSON)

		assert.Equal(t, expected.inputJSON, string(givenInputJSON))

		res.WriteHeader(expected.responseCode)
		res.Write([]byte(expected.responseJSON))
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

	t.Run("Credits HTTP error", func(t *testing.T) {
		// mock openai API
		ts := startTestServer(
			t,
			[]testRequest{
				{
					"",
					`{
									"error": {
										"detail": "The provided API key does not exist",
										"status": 403,
										"title": "Forbidden",
										"code": 104
									}
								}`,
					http.StatusForbidden,
				},
			},
		)
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
		mocks.AssertError(slackClient, message, "API call returned an err: 403")

		actual := commands.Run(message)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, actual)
	})

	t.Run("Credits HTTP works", func(t *testing.T) {
		// mock openai API
		ts := startTestServer(
			t,
			[]testRequest{
				{
					"",
					`{
									"current_balance": 998060,
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
									"transactions": "https://atlas.ripe.net/api/v2/credits/transactions/"
								}`,
					http.StatusOK,
				},
			},
		)
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
}
