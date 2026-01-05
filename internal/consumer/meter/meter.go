package meter

import (
	"context"

	"github.com/realfatcat/raspiweather/internal/consumer"
	"github.com/realfatcat/raspiweather/internal/types"
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
				m.metrics.RecordTemperature(ctx, sd.Temperature)
				m.metrics.RecordHumidity(ctx, sd.Humidity)
				m.metrics.RecordPressure(ctx, sd.Pressure)
			}
		}
	}()
}
