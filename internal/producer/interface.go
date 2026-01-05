package producer

import "github.com/realfatcat/raspiweather/internal/types"

type Bus interface {
	Read() (types.WeatherData, error)
	Close() error
}
