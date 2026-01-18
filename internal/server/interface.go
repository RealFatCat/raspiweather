package server

import "github.com/realfatcat/raspiweather/pkg/types"

type Producer interface {
	ReadWeatherData() ([]types.WeatherData, error)
}
