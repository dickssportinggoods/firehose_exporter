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

var _ = Describe("CounterEventsCollector", func() {
	var (
		namespace               string
		environment             string
		metricsStore            *metrics.Store
		metricsExpiration       time.Duration
		metricsCleanupInterval  time.Duration
		metricsCustomUuidOrigin string
		deploymentFilter        *filters.DeploymentFilter
		eventFilter             *filters.EventFilter
		counterEventsCollector  *CounterEventsCollector

		counterEventsCollectorDesc *prometheus.Desc
	)

	BeforeEach(func() {
		namespace = "test_exporter"
		environment = "test_environment"
		deploymentFilter = filters.NewDeploymentFilter([]string{})
		eventFilter, _ = filters.NewEventFilter([]string{})
		metricsCustomUuidOrigin = ""
		metricsStore = metrics.NewStore(metricsExpiration, metricsCleanupInterval, deploymentFilter, eventFilter, metricsCustomUuidOrigin)

		counterEventsCollectorDesc = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "counter_event", "collector"),
			"Cloud Foundry Firehose counter metrics collector.",
			nil,
			prometheus.Labels{"environment": environment},
		)
	})

	JustBeforeEach(func() {
		counterEventsCollector = NewCounterEventsCollector(namespace, environment, metricsStore)
	})

	Describe("Describe", func() {
		var (
			descriptions chan *prometheus.Desc
		)

		BeforeEach(func() {
			descriptions = make(chan *prometheus.Desc)
		})

		JustBeforeEach(func() {
			go counterEventsCollector.Describe(descriptions)
		})

		It("returns a counter_event_collector metric description", func() {
			Eventually(descriptions).Should(Receive(Equal(counterEventsCollectorDesc)))
		})
	})

	Describe("Collect", func() {
		var (
			boshDeployment = "fake-deployment-name"
			boshJob        = "fake-job-name"
			boshIndex      = "0"
			boshIP         = "1.2.3.4"

			counterEvent1Origin               = "fake.origin"
			counterEvent1OriginNameNormalized = "fake_origin"
			counterEvent1OriginDescNormalized = "fake-origin"
			counterEvent1Name                 = "FakeCounterEvent1"
			counterEvent1NameNormalized       = "fake_counter_event_1"
			counterEvent1DescNormalized       = "FakeCounterEvent1"
			counterEvent1Delta                = uint64(5)
			counterEvent1Total                = uint64(1000)

			counterEvent2Origin               = "p.fake.origin"
			counterEvent2OriginNameNormalized = "p_fake_origin"
			counterEvent2OriginDescNormalized = "p-fake-origin"
			counterEvent2Name                 = "/p.fake/CounterEvent2"
			counterEvent2NameNormalized       = "p_fake_counter_event_2"
			counterEvent2DescNormalized       = "/p-fake/CounterEvent2"
			counterEvent2Delta                = uint64(10)
			counterEvent2Total                = uint64(2000)

			counterEventsChan  chan prometheus.Metric
			totalCounterEvent1 prometheus.Metric
			deltaCounterEvent1 prometheus.Metric
			totalCounterEvent2 prometheus.Metric
			deltaCounterEvent2 prometheus.Metric

			tag1Name           = "tag1"
			tag1NameNormalized = "tag1"
			tag1Value          = "fakeTag1"

			tag2Name           = "tag2"
			tag2NameNormalized = "tag2"
			tag2Value          = "fakeTag2"
		)

		BeforeEach(func() {
			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(counterEvent1Origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEvent1Name),
						Delta: proto.Uint64(counterEvent1Delta),
						Total: proto.Uint64(counterEvent1Total),
					},
					Tags: map[string]string{
						tag1Name: tag1Value,
					},
				},
			)

			metricsStore.AddMetric(
				&events.Envelope{
					Origin:     proto.String(counterEvent2Origin),
					EventType:  events.Envelope_CounterEvent.Enum(),
					Timestamp:  proto.Int64(time.Now().Unix() * 1000),
					Deployment: proto.String(boshDeployment),
					Job:        proto.String(boshJob),
					Index:      proto.String(boshIndex),
					Ip:         proto.String(boshIP),
					CounterEvent: &events.CounterEvent{
						Name:  proto.String(counterEvent2Name),
						Delta: proto.Uint64(counterEvent2Delta),
						Total: proto.Uint64(counterEvent2Total),
					},
					Tags: map[string]string{
						tag2Name: tag2Value,
					},
				},
			)

			counterEventsChan = make(chan prometheus.Metric)

			totalCounterEvent1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", counterEvent1OriginNameNormalized+"_"+counterEvent1NameNormalized+"_total"),
					fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event from '%s'.", counterEvent1DescNormalized, counterEvent1OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", tag1NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.CounterValue,
				float64(counterEvent1Total),
				counterEvent1Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				tag1Value,
			)

			deltaCounterEvent1 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", counterEvent1OriginNameNormalized+"_"+counterEvent1NameNormalized+"_delta"),
					fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event from '%s'.", counterEvent1DescNormalized, counterEvent1OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", tag1NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.GaugeValue,
				float64(counterEvent1Delta),
				counterEvent1Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				tag1Value,
			)

			totalCounterEvent2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", counterEvent2OriginNameNormalized+"_"+counterEvent2NameNormalized+"_total"),
					fmt.Sprintf("Cloud Foundry Firehose '%s' total counter event from '%s'.", counterEvent2DescNormalized, counterEvent2OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", tag2NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.CounterValue,
				float64(counterEvent2Total),
				counterEvent2Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				tag2Value,
			)

			deltaCounterEvent2 = prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "counter_event", counterEvent2OriginNameNormalized+"_"+counterEvent2NameNormalized+"_delta"),
					fmt.Sprintf("Cloud Foundry Firehose '%s' delta counter event from '%s'.", counterEvent2DescNormalized, counterEvent2OriginDescNormalized),
					[]string{"origin", "bosh_deployment", "bosh_job_name", "bosh_job_id", "bosh_job_ip", tag2NameNormalized},
					prometheus.Labels{"environment": environment},
				),
				prometheus.GaugeValue,
				float64(counterEvent2Delta),
				counterEvent2Origin,
				boshDeployment,
				boshJob,
				boshIndex,
				boshIP,
				tag2Value,
			)
		})

		JustBeforeEach(func() {
			go counterEventsCollector.Collect(counterEventsChan)
		})

		It("returns a counter_event_fake_origin_fake_counter_event_1_total metric", func() {
			Eventually(counterEventsChan).Should(Receive(PrometheusMetric(totalCounterEvent1)))
		})

		It("returns a counter_event_fake_origin_fake_counter_event_1_delta metric", func() {
			Eventually(counterEventsChan).Should(Receive(PrometheusMetric(deltaCounterEvent1)))
		})

		It("returns a counter_event_fake_origin_fake_counter_event_2_total metric", func() {
			Eventually(counterEventsChan).Should(Receive(PrometheusMetric(totalCounterEvent2)))
		})

		It("returns a counter_event_fake_origin_fake_counter_event_2_delta metric", func() {
			Eventually(counterEventsChan).Should(Receive(PrometheusMetric(deltaCounterEvent2)))
		})

		Context("when there is no counter metrics", func() {
			BeforeEach(func() {
				metricsStore.FlushCounterEvents()
			})

			It("does not return any metric", func() {
				Consistently(counterEventsChan).ShouldNot(Receive())
			})
		})
	})
})
