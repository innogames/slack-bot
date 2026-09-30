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
	"github.com/prometheus/client_golang/prometheus"
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

	// custom collector, e.g. provided by a plugin
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{Name: "plugin_test_gauge"})
	gauge.Set(42)
	RegisterCollector(gauge)

	// a collector with the same name can't be registered twice: it's skipped with a warning
	RegisterCollector(prometheus.NewGauge(prometheus.GaugeOpts{Name: "plugin_test_gauge"}))
	defer func() {
		customCollectors = nil
	}()

	InitMetrics(cfg, ctx)
	time.Sleep(time.Millisecond * 10)

	resp, err := http.Get("http://" + metricsPort + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	content, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(content), "slack_bot_test_value 500")
	assert.Contains(t, string(content), "plugin_test_gauge 42")
}

// get a random free port on the host
func getPort() string {
	lc := &net.ListenConfig{}
	l, _ := lc.Listen(context.Background(), "tcp4", "localhost:0")
	defer l.Close()

	return l.Addr().String()
}
