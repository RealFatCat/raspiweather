package server

import "github.com/realfatcat/raspiweather/internal/producer"

type Producer interface {
	ReadWeatherData() producer.ReadWeatherResult
}

type BacklightToggler interface {
	ToggleBacklight() error
}
