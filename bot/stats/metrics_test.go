package stats

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	ctx := util.NewServerContext()
	defer ctx.StopTheWorld()

	metricsPort := getPort()

	cfg := config.Config{
		Metrics: config.Metrics{
			PrometheusListener: metricsPort,
		},
	}

	Set("test_value", 500)

	InitMetrics(cfg, ctx)
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+metricsPort+"/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	content, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(content), "slack_bot_test_value 500")
}

// get a random free port on the host
func getPort() string {
	lc := &net.ListenConfig{}
	l, _ := lc.Listen(context.Background(), "tcp4", "localhost:0")
	defer l.Close()

	return l.Addr().String()
}
