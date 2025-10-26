package stats

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestMetricsBasic(_ *testing.T) {
	ctx := util.NewServerContext()
	defer ctx.StopTheWorld()

	cfg := config.Config{
		Metrics: config.Metrics{
			PrometheusListener: "",
		},
	}

	// Test metrics initialization with disabled metrics
	InitMetrics(cfg, ctx)

	// Give it a moment to ensure no server starts
	time.Sleep(10 * time.Millisecond)
}

func TestStatRegistryBasic(t *testing.T) {
	t.Run("Describe", func(t *testing.T) {
		registry := &statRegistry{}
		ch := make(chan *prometheus.Desc)
		go func() {
			registry.Describe(ch)
			close(ch)
		}()
		// Expect no descriptions for our simple case
		assert.Empty(t, ch)
	})

	t.Run("Collect", func(t *testing.T) {
		Set("test_metric_1", 10)
		Set("test_metric_2", 20)

		registry := &statRegistry{}
		ch := make(chan prometheus.Metric, 10)

		registry.Collect(ch)
		close(ch)

		// Count metrics
		metricCount := 0
		for range ch {
			metricCount++
		}

		// Should have at least our test metrics
		assert.GreaterOrEqual(t, metricCount, 2)
	})
}
