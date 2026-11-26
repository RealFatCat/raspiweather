package producer

import "github.com/realfatcat/raspiweather/pkg/types"

// SensorError describes a failed sensor read.
type SensorError struct {
	SensorID string
	Err      error
}

func (e *SensorError) Error() string {
	if e.Err == nil {
		return e.SensorID
	}
	return e.SensorID + ": " + e.Err.Error()
}

// ReadWeatherResult holds the outcome of a multi-sensor read.
type ReadWeatherResult struct {
	Data   []types.WeatherData
	Errors []SensorError
}

