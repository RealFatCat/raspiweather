package producer

import (
	"context"
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

// ReadWeatherData reads from all sensors and returns a slice of weather data
func (p *MultiProducer) ReadWeatherData() ([]types.WeatherData, error) {
	results := make([]types.WeatherData, 0, len(p.sensors))
	for _, sensor := range p.sensors {
		data, err := sensor.Bus.Read()
		if err != nil {
			slog.Error("reading sensor data", "sensor_id", sensor.ID, "error", err)
			continue
		}
		data.SensorID = sensor.ID
		results = append(results, data)
	}

	return results, nil
}
