package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorsMetrics struct {
	temperature prometheus.Gauge
	humidity    prometheus.Gauge
	pressure    prometheus.Gauge
}

func NewSensorsMetrics() *SensorsMetrics {
	return &SensorsMetrics{
		temperature: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "sensor_temperature",
			Help:        "Current temperature.",
			ConstLabels: prometheus.Labels{"unit": "celsius"},
		}),
		humidity: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "sensor_humidity",
			Help:        "Current humidity.",
			ConstLabels: prometheus.Labels{"unit": "percent"},
		}),
		pressure: promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "sensor_pressure",
			Help:        "Current pressure.",
			ConstLabels: prometheus.Labels{"unit": "hPa"},
		}),
	}
}

func (s *SensorsMetrics) RecordTemperature(_ context.Context, num float64) {
	s.temperature.Set(num)
}

func (s *SensorsMetrics) RecordHumidity(_ context.Context, num float64) {
	s.humidity.Set(num)
}

func (s *SensorsMetrics) RecordPressure(_ context.Context, num float64) {
	s.pressure.Set(num)
}
