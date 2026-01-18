package meter

import (
	"context"

	"github.com/realfatcat/raspiweather/internal/consumer"
	"github.com/realfatcat/raspiweather/pkg/types"
)

type Meter struct {
	metrics Metrics
}

var _ consumer.Consumer = (*Meter)(nil)

func New(metrics Metrics) *Meter {
	return &Meter{
		metrics: metrics,
	}
}

func (m *Meter) Consume(ctx context.Context, ch <-chan types.WeatherData) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sd, ok := <-ch:
				if !ok {
					// Channel closed, exit gracefully
					return
				}
				sensorID := sd.SensorID
				if sensorID == "" {
					sensorID = "default"
				}
				m.metrics.RecordTemperature(ctx, sensorID, sd.Temperature)
				m.metrics.RecordHumidity(ctx, sensorID, sd.Humidity)
				m.metrics.RecordPressure(ctx, sensorID, sd.Pressure)
			}
		}
	}()
}
