package collectors_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/bosh-prometheus/firehose_exporter/filters"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	. "github.com/bosh-prometheus/firehose_exporter/collectors"
	. "github.com/bosh-prometheus/firehose_exporter/utils/test_matchers"
)

var _ = Describe("ValueMetricsCollector", func() {
	var (
		namespace               string
		environment             string
		metricsStore            *metrics.Store
		metricsExpiration       time.Duration
		metricsCleanupInterval  time.Duration
		metricsCustomUuidOrigin string
		deploymentFilter        *filters.DeploymentFilter
		eventFilter             *filters.EventFilter
		valueMetricsCollector   *ValueMetricsCollector

		valueMetricsCollectorDesc *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		environment = "test_environment"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsCustomUuidOrigin = ""
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter, metricsCustomUuidOrigin)

		valueMetricsCollectorDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "value_metric", "collector"),
			"Cloud Foundry Firehose value metrics collector.",
			nil,
			prometheus.Labels{"environment": environment},
		)
	})

	JustBeforeEach(func() {
		valueMetricsCollector = NewValueMetricsCollector(namespace, environment, metricsStore)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go valueMetricsCollector.Describe(descriptions)
		})

		It("returns a value_metric_collector metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(valueMetricsCollectorDesc)))
		})
	})

	Describe("Collect", func() {
		var (
			boshDeployment = "fake-deployment-name"
			boshJob        = "fake-job-name"
			boshIndex      = "0"
			boshIP         = "1.2.3.4"

			valueMetric1Origin               = "fake.origin"
			valueMetric1OriginNameNormalized = "fake_origin"
			valueMetric1OriginDescNormalized = "fake-origin"
			valueMetric1Name                 = "FakeValueMetric1"
			valueMetric1NameNormalized       = "fake_value_metric_1"
			valueMetric1DescNormalized       = "FakeValueMetric1"
			valueMetric1Value                = float64(2000)
			valueMetric1Unit                 = "kb"

			valueMetric2Origin               = "p.fake.origin"
			valueMetric2OriginNameNormalized = "p_fake_origin"
			valueMetric2OriginDescNormalized = "p-fake-origin"
			valueMetric2Name                 = "/p.fake/ValueMetric2"
			valueMetric2NameNormalized       = "p_fake_value_metric_2"
			valueMetric2DescNormalized       = "/p-fake/ValueMetric2"
			valueMetric2Value                = float64(15)
			valueMetric2Unit                 = "count"

			valueMetricsChan chan prometheus.Metric
			valueMetric1     prometheus.Metric
			valueMetric2     prometheus.Metric

			tag1Name           = "tag1"
			tag1NameNormalized = "tag_1"
			tag1Value          = "fakeTag1"

			tag2Name           = "tag-2"
			tag2NameNormalized = "tag_2"
			tag2Value          = "fakeTag2"
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(valueMetric1Origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetric1Name),
						Value: proto.Float64(valueMetric1Value),
						Unit:  proto.String(valueMetric1Unit),
					},
					Tags: map[string]string{
						tag1Name: tag1Value,
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(valueMetric2Origin),
					EventType:  events.Envelope_ValueMetric.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					ValueMetric: &events.ValueMetric{
						Name:  proto.String(valueMetric2Name),
						Value: proto.Float64(valueMetric2Value),
						Unit:  proto.String(valueMetric2Unit),
					},
					Tags: map[string]string{
						tag2Name: tag2Value,
					},
				},
			)

			valueMetricsChan = make(chan prometheus.Metric)

			valueMetric1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "value_metric", valueMetric1OriginNameNormalized+"_"+valueMetric1NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' value metric from '%s'.", valueMetric1DescNormalized, valueMetric1OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "unit", tag1NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.GaugeValue,
				valueMetric1Value,
				valueMetric1Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				valueMetric1Unit,
				tag1Value,
			)

			valueMetric2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "value_metric", valueMetric2OriginNameNormalized+"_"+valueMetric2NameNormalized),
					fmt.Sprintf("Cloud Foundry Firehose '%s' value metric from '%s'.", valueMetric2DescNormalized, valueMetric2OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", "unit", tag2NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.GaugeValue,
				valueMetric2Value,
				valueMetric2Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				valueMetric2Unit,
				tag2Value,
			)
		})

		JustBeforeEach(func() {
			go valueMetricsCollector.Collect(valueMetricsChan)
		})

		It("returns a value_metric_fake_origin_fake_value_metric_1 metric", func() {
			Eventually(valueMetricsChan).Should(Receive(PrometheusMetric(valueMetric1)))
		})

		It("returns a value_metric_fake_origin_fake_value_metric_2 metric", func() {
			Eventually(valueMetricsChan).Should(Receive(PrometheusMetric(valueMetric2)))
		})

		Context("when there is no value metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushValueMetrics()
			})

			It("does not return any metric", func() {
				Consistently(valueMetricsChan).ShouldNot(Receive())
			})
		})
	})
})
