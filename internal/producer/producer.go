package producer

import (
	"context"
	"log/slog"
	"time"

	"github.com/realfatcat/raspiweather/internal/types"
)

type Producer struct {
	interval time.Duration
	bus      Bus
}

func New(bus Bus, interval time.Duration) *Producer {
	return &Producer{
		interval: interval,
		bus:      bus,
	}
}

func (p *Producer) Produce(ctx context.Context) <-chan types.WeatherData {
	ch := make(chan types.WeatherData, 1)

	go func() {
		defer close(ch)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		// initial read
		if err := p.produce(ctx, ch); err != nil {
			slog.Error("reading sensor data", "error", err)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.produce(ctx, ch); err != nil {
					slog.Error("reading sensor data", "error", err)
				}
			}
		}
	}()

	return ch
}

func (p *Producer) produce(ctx context.Context, ch chan<- types.WeatherData) error {
	data, err := p.ReadWeatherData()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch <- data:
	}
	return nil
}

func (p *Producer) ReadWeatherData() (types.WeatherData, error) {
	return p.bus.Read()
}
