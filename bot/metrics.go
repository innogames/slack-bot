package bot

import (
	"net/http"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type statRegistry struct{}

// Describe returns all descriptions of the collector.
func (c *statRegistry) Describe(_ chan<- *prometheus.Desc) {
	// unused in our simple case...
}

// Collect returns the current state of all metrics of our slack-bot stats
func (c *statRegistry) Collect(ch chan<- prometheus.Metric) {
	for _, key := range stats.GetKeys() {
		metric := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "slack_bot",
			Name:      strings.ReplaceAll(key, "-", "_"),
		})
		value, _ := stats.Get(key)
		metric.Set(float64(value))
		metric.Collect(ch)
	}
}

func initMetrics(cfg config.Config, ctx *util.ServerContext) {
	if cfg.Metrics.PrometheusListener == "" {
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		&statRegistry{},
		collectors.NewGoCollector(),
	)

	go func() {
		ctx.RegisterChild()
		defer ctx.ChildDone()

		log.Infof("Init prometheus handler on http://%s/metrics", cfg.Metrics.PrometheusListener)

		server := &http.Server{
			Addr:              cfg.Metrics.PrometheusListener,
			ReadHeaderTimeout: 3 * time.Second,
		}

		http.Handle(
			"/metrics", promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{},
			),
		)

		go func() {
			_ = server.ListenAndServe()
		}()

		<-ctx.Done()
		_ = server.Shutdown(ctx)
	}()
}
