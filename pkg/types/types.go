package types

type WeatherData struct {
	SensorID    string  `json:"sensor_id,omitempty"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
}
