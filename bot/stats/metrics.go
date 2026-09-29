package stats

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type statRegistry struct{}

var (
	customCollectors     []prometheus.Collector
	customCollectorsLock sync.Mutex
)

// RegisterCollector adds a custom prometheus collector to the "/metrics" endpoint, e.g. to provide metrics of a plugin.
// It has to be called before the bot is started, e.g. in the Setup function of a plugin.
func RegisterCollector(collector prometheus.Collector) {
	customCollectorsLock.Lock()
	defer customCollectorsLock.Unlock()

	customCollectors = append(customCollectors, collector)
}

// Describe returns all descriptions of the collector.
func (c *statRegistry) Describe(_ chan<- *prometheus.Desc) {
	// unused in our simple case...
}

// Collect returns the current state of all metrics of our slack-bot stats
func (c *statRegistry) Collect(ch chan<- prometheus.Metric) {
	for _, key := range GetKeys() {
		metric := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "slack_bot",
			Name:      strings.ReplaceAll(key, "-", "_"),
		})
		value, _ := Get(key)
		metric.Add(float64(value))
		metric.Collect(ch)
	}
}

func InitMetrics(cfg config.Config, ctx *util.ServerContext) {
	if !cfg.Metrics.IsEnabled() {
		// prometheus is disabled...skip here
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		&statRegistry{},
		collectors.NewGoCollector(),
	)

	customCollectorsLock.Lock()
	for _, collector := range customCollectors {
		if err := registry.Register(collector); err != nil {
			log.Warnf("Failed to register custom prometheus collector: %s", err)
		}
	}
	customCollectorsLock.Unlock()

	ctx.Go(func() {
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
			err := server.ListenAndServe()
			if err != nil {
				log.Warnf("Failed to start prometheus server: %s", err)
			}
		}()

		<-ctx.Done()
		_ = server.Shutdown(ctx)
	})
}
