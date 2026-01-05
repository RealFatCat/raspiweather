package server

import "github.com/realfatcat/raspiweather/internal/types"

type Producer interface {
	ReadWeatherData() (types.WeatherData, error)
}
