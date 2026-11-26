package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

type SensorsMetrics struct {
	temperature *prometheus.GaugeVec
	humidity    *prometheus.GaugeVec
	pressure    *prometheus.GaugeVec
}

// NewSensorsMetrics creates and registers the sensors metrics with reg.
// Registration errors are returned, not panicked; on failure already-registered
// collectors are unregistered so the caller can retry safely.
func NewSensorsMetrics(reg prometheus.Registerer) (s *SensorsMetrics, err error) {
	var registered []prometheus.Collector
	defer func() {
		if err != nil {
			for _, c := range registered {
				reg.Unregister(c)
			}
		}
	}()

	var temperature *prometheus.GaugeVec
	temperature, err = gaugeVec(reg, "sensor_temperature", "Current temperature.", []string{"sensor_id", "unit"})
	if err != nil {
		return nil, err
	}
	registered = append(registered, temperature)

	var humidity *prometheus.GaugeVec
	humidity, err = gaugeVec(reg, "sensor_humidity", "Current humidity.", []string{"sensor_id", "unit"})
	if err != nil {
		return nil, err
	}
	registered = append(registered, humidity)

	var pressure *prometheus.GaugeVec
	pressure, err = gaugeVec(reg, "sensor_pressure", "Current pressure.", []string{"sensor_id", "unit"})
	if err != nil {
		return nil, err
	}
	registered = append(registered, pressure)

	return &SensorsMetrics{
		temperature: temperature,
		humidity:    humidity,
		pressure:    pressure,
	}, nil
}

func (s *SensorsMetrics) RecordTemperature(_ context.Context, sensorID string, num float64) {
	s.temperature.WithLabelValues(sensorID, "celsius").Set(num)
}

func (s *SensorsMetrics) RecordHumidity(_ context.Context, sensorID string, num float64) {
	s.humidity.WithLabelValues(sensorID, "percent").Set(num)
}

func (s *SensorsMetrics) RecordPressure(_ context.Context, sensorID string, num float64) {
	s.pressure.WithLabelValues(sensorID, "hPa").Set(num)
}
