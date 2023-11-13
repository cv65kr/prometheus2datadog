package internal

import (
	"math"
	"strings"

	dto "github.com/prometheus/client_model/go"
	"go.uber.org/zap"
)

type converter struct {
	datadogClient *datadogClient
	logger        *zap.Logger
	config        Config
	previous      map[string]int64
}

func NewConverter(datadogClient *datadogClient, logger *zap.Logger, config Config) converter {
	return converter{
		datadogClient: datadogClient,
		logger:        logger,
		config:        config,
		previous:      make(map[string]int64),
	}
}

func (c converter) Convert(mf map[string]*dto.MetricFamily) {
	for name, metric := range mf {
		for _, prefix := range c.config.ExcludedMetrics {
			if strings.HasPrefix(name, prefix) {
				c.logger.Debug("Skipping sending metric", zap.String("metric", name))
				continue
			}
		}

		c.logger.Info("Sending metric to Datadog", zap.String("metric", name))
		switch metric.GetType() {
		case dto.MetricType_COUNTER:
			for _, m := range metric.GetMetric() {
				v := m.GetCounter().GetValue()
				if math.IsInf(v, 0) || math.IsNaN(v) {
					continue
				}

				c.counter(name, v, m.GetLabel(), "Error during sending COUNTER metric")
			}
			break
		case dto.MetricType_GAUGE:
			for _, m := range metric.GetMetric() {
				v := m.GetGauge().GetValue()

				err := c.datadogClient.statsd.Gauge(name, v, c.getTags(m.GetLabel()), 1)
				if err != nil {
					c.logger.Error("Error during sending GAUGE metric", zap.Error(err), zap.String("metric", name))
				}
			}
			break
		case dto.MetricType_HISTOGRAM:
			for _, m := range metric.GetMetric() {
				histogram := m.GetHistogram()
				tags := c.getTags(m.GetLabel())

				// https://docs.datadoghq.com/integrations/guide/prometheus-metrics/?tab=latestversion#histogram
				err := c.datadogClient.statsd.Count(name+".count", int64(histogram.GetSampleCount()), tags, 1)
				if err != nil {
					c.logger.Error("Error during sending HISTOGRAM (.count) metric", zap.Error(err), zap.String("metric", name))
				}
				err = c.datadogClient.statsd.Count(name+".sum", int64(histogram.GetSampleSum()), tags, 1)
				if err != nil {
					c.logger.Error("Error during sending HISTOGRAM (.sum) metric", zap.Error(err), zap.String("metric", name))
				}

				for _, b := range histogram.GetBucket() {
					err = c.datadogClient.statsd.Count(name+".bucket", int64(*b.CumulativeCount), tags, 1)
					if err != nil {
						c.logger.Error("Error during sending HISTOGRAM (.bucket) metric", zap.Error(err), zap.String("metric", name))
					}
				}
			}
			break
		case dto.MetricType_SUMMARY:
			for _, m := range metric.GetMetric() {
				summary := m.GetSummary()
				tags := c.getTags(m.GetLabel())

				// https://docs.datadoghq.com/integrations/guide/prometheus-metrics/?tab=latestversion#summary
				err := c.datadogClient.statsd.Count(name+".count", int64(summary.GetSampleCount()), tags, 1)
				if err != nil {
					c.logger.Error("Error during sending SUMMARY (.count) metric", zap.Error(err), zap.String("metric", name))
				}

				err = c.datadogClient.statsd.Count(name+".sum", int64(summary.GetSampleSum()), tags, 1)
				if err != nil {
					c.logger.Error("Error during sending SUMMARY (.sum) metric", zap.Error(err), zap.String("metric", name))
				}

				for _, q := range summary.GetQuantile() {
					err = c.datadogClient.statsd.Gauge(name+".quantile", q.GetQuantile(), tags, 1)
					if err != nil {
						c.logger.Error("Error during sending SUMMARY (.quantile) metric", zap.Error(err), zap.String("metric", name))
					}
				}
			}
		}
	}
}

func (c converter) counter(name string, value float64, tags []*dto.LabelPair, errorMessage string) {
	// Get previous value
	var previousVal int64 = 0
	if _, ok := c.previous[name]; ok {
		previousVal = c.previous[name]
	}

	currentVal := int64(value)
	c.previous[name] = currentVal

	if previousVal != 0 && currentVal != 0 {
		currentVal = max(0, currentVal-previousVal)
	}

	err := c.datadogClient.statsd.Count(name, currentVal, c.getTags(tags), 1)
	if err != nil {
		c.logger.Error(errorMessage, zap.Error(err), zap.String("metric", name))
	}
}

func (c converter) getTags(labels []*dto.LabelPair) []string {
	var tags []string
	for _, l := range labels {
		tags = append(tags, l.GetName()+":"+l.GetValue())
	}

	return tags
}
