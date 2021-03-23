package collectors

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"

	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/utils"
)

type ValueMetricsCollector struct {
	namespace                 string
	environment               string
	metricsStore              *metrics.Store
	valueMetricsCollectorDesc *prometheus.Desc
}

func NewValueMetricsCollector(
	namespace string,
	environment string,
	metricsStore *metrics.Store,
) *ValueMetricsCollector {
	valueMetricsCollectorDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, value_metrics_subsystem, "collector"),
		"Cloud Foundry Firehose value metrics collector.",
		nil,
		prometheus.Labels{"environment": environment},
	)

	return &ValueMetricsCollector{
		namespace:                 namespace,
		environment:               environment,
		metricsStore:              metricsStore,
		valueMetricsCollectorDesc: valueMetricsCollectorDesc,
	}
}

func (c ValueMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	for _, valueMetric := range c.metricsStore.GetValueMetrics() {
		metricName := utils.NormalizeName(valueMetric.Origin) + "_" + utils.NormalizeName(valueMetric.Name)
		if utils.IsValidUuid(valueMetric.Origin) {
			metricName = "custom_" + utils.NormalizeName(valueMetric.Name)
		}

		constLabels := []string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "unit"}
		labelValues := []string{valueMetric.Origin, valueMetric.Deployment, valueMetric.Job, valueMetric.Index, valueMetric.IP, valueMetric.Unit}

		for k, v := range valueMetric.Tags {
			if k == "unit" {
				if v != valueMetric.Unit && valueMetric.Unit == "" {
					labelValues[5] = v
					// Dont add a duplicate unit label from tags and correctly set valueMetric.Unit if unset.
					// This does not solve all cases where a Tag may match a constLabel however it resolves
					// https://github.com/bosh-prometheus/firehose_exporter/issues/59 logstore discards with
					// limited performance impact. log_store is sending metrics with units in the tags but not
					// correctly populating valueMetric.Unit - likely needs to be fixed in upstream prom-scraper
					continue
				}
			}
			constLabels = append(constLabels, utils.NormalizeName(k))
			labelValues = append(labelValues, v)
		}

		vm, err := prometheus.NewConstMetric(
			prometheus.NewDesc(
				prometheus.BuildFQName(c.namespace, value_metrics_subsystem, metricName),
				fmt.Sprintf("Cloud Foundry Firehose '%s' value metric from '%s'.", utils.NormalizeNameDesc(valueMetric.Name), utils.NormalizeOriginDesc(valueMetric.Origin)),
				constLabels,
				prometheus.Labels{"environment": c.environment},
			),
			prometheus.GaugeValue,
			float64(valueMetric.Value),
			labelValues...,
		)

		if err != nil {
			log.Errorf("Value Metric `%s` from `%s` discarded: %s", valueMetric.Name, valueMetric.Origin, err)
			continue
		}
		ch <- vm
	}
}

func (c ValueMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.valueMetricsCollectorDesc
}
