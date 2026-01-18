package producer

import (
	"context"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type Bus interface {
	Read() (types.WeatherData, error)
	Close() error
}

// ProducerInterface defines the interface for weather data producers
type ProducerInterface interface {
	Produce(ctx context.Context) <-chan types.WeatherData
	ReadWeatherData() (types.WeatherData, error)
}
