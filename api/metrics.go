package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	mutex *sync.Mutex
	reg *prometheus.Registry
	collectors map[string]prometheus.Collector
	labels map[string][]string
}

var metricsSingleton = &Metrics{
	mutex: &sync.Mutex{},
	reg: prometheus.NewRegistry(),
	collectors: map[string]prometheus.Collector{},
	labels: map[string][]string{},
}

func init() {
	start := time.Now()
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			<-ticker.C
			Measure("uptime", nil, time.Since(start).Seconds())
		}
	}()
	metricsSingleton.reg.Register(collectors.NewGoCollector())
	metricsSingleton.reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

func MetricsEndpoint() http.HandlerFunc {
	m := metricsSingleton
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{Registry: m.reg}).(http.HandlerFunc)
}

func Measure(name string, labels map[string]string, value float64) {
	m := metricsSingleton
	m.mutex.Lock()
	defer m.mutex.Unlock()
	collector, ok := m.collectors[name]
	if !ok {
		var labelKeys []string
		if labels == nil || len(labels) == 0 {
			collector = prometheus.NewGauge(prometheus.GaugeOpts{Name: name})
		} else {
			for k := range labels {
				labelKeys = append(labelKeys, k)
			}
			collector = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: name}, labelKeys)
		}
		m.collectors[name] = collector
		m.labels[name] = labelKeys
		m.reg.Register(collector)
	}
	labelKeys := m.labels[name]
	if labelKeys == nil {
		gauge, ok := collector.(prometheus.Gauge)
		if !ok {
			return
		}
		gauge.Set(value)
	} else {
		gauge, ok := collector.(*prometheus.GaugeVec)
		if !ok {
			return
		}
		labelVals := make([]string, len(labelKeys))
		for i, k := range labelKeys {
			labelVals[i] = labels[k]
		}
		gauge.WithLabelValues(labelVals...).Set(value)
	}
}

func Count(name string, labels map[string]string) {
	m := metricsSingleton
	m.mutex.Lock()
	defer m.mutex.Unlock()
	collector, ok := m.collectors[name]
	if !ok {
		var labelKeys []string
		if labels == nil || len(labels) == 0 {
			collector = prometheus.NewCounter(prometheus.CounterOpts{Name: name})
		} else {
			for k := range labels {
				labelKeys = append(labelKeys, k)
			}
			collector = prometheus.NewCounterVec(prometheus.CounterOpts{Name: name}, labelKeys)
		}
		m.collectors[name] = collector
		m.labels[name] = labelKeys
		m.reg.Register(collector)
	}
	labelKeys := m.labels[name]
	if labelKeys == nil {
		counter, ok := collector.(prometheus.Counter)
		if !ok {
			return
		}
		counter.Inc()
	} else {
		counter, ok := collector.(*prometheus.CounterVec)
		if !ok {
			return
		}
		labelVals := make([]string, len(labelKeys))
		for i, k := range labelKeys {
			labelVals[i] = labels[k]
		}
		counter.WithLabelValues(labelVals...).Inc()
	}
}
