package producer

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/realfatcat/raspiweather/pkg/types"
)

// SensorConfig represents a sensor configuration with its ID and bus
type SensorConfig struct {
	ID  string
	Bus Bus
}

// MultiProducer handles multiple sensors and produces data from all of them
type MultiProducer struct {
	interval time.Duration
	sensors  []SensorConfig
}

// NewMultiProducer creates a new MultiProducer with multiple sensor configurations
func NewMultiProducer(sensors []SensorConfig, interval time.Duration) *MultiProducer {
	return &MultiProducer{
		interval: interval,
		sensors:  sensors,
	}
}

// Produce produces weather data from all sensors at the specified interval
func (p *MultiProducer) Produce(ctx context.Context) <-chan types.WeatherData {
	ch := make(chan types.WeatherData, len(p.sensors))

	go func() {
		defer close(ch)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		// initial read from all sensors
		p.produceAll(ctx, ch)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.produceAll(ctx, ch)
			}
		}
	}()

	return ch
}

// produceAll reads from all sensors and sends data to the channel
func (p *MultiProducer) produceAll(ctx context.Context, ch chan<- types.WeatherData) {
	var wg sync.WaitGroup
	for _, sensor := range p.sensors {
		wg.Go(func() {
			data, err := sensor.Bus.Read()
			if err != nil {
				slog.Error("reading sensor data", "sensor_id", sensor.ID, "error", err)
				return
			}
			data.SensorID = sensor.ID

			select {
			case <-ctx.Done():
				return
			case ch <- data:
			}
		})
	}
	wg.Wait()
}

// ReadWeatherData reads from all sensors. Partial failures are logged and
// recorded in the returned Result; they do not fail the whole call.
func (p *MultiProducer) ReadWeatherData() ReadWeatherResult {
	if len(p.sensors) == 0 {
		return ReadWeatherResult{Errors: []SensorError{{SensorID: "", Err: errors.New("no sensors configured")}}}
	}

	res := ReadWeatherResult{
		Data:   make([]types.WeatherData, 0, len(p.sensors)),
		Errors: make([]SensorError, 0, len(p.sensors)),
	}
	for _, sensor := range p.sensors {
		data, err := sensor.Bus.Read()
		if err != nil {
			slog.Error("reading sensor data", "sensor_id", sensor.ID, "error", err)
			res.Errors = append(res.Errors, SensorError{SensorID: sensor.ID, Err: err})
			continue
		}
		data.SensorID = sensor.ID
		res.Data = append(res.Data, data)
	}
	return res
}
