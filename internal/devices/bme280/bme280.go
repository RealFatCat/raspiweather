package bme280

import (
	"fmt"

	"golang.org/x/exp/io/i2c"

	"github.com/quhar/bme280"

	"github.com/realfatcat/raspiweather/internal/types"
)

const (
	DefaultDevPath    = "/dev/i2c-1"
	DefaultI2CAddress = bme280.I2CAddr
)

type Bus struct {
	device *i2c.Device
	bme280 *bme280.BME280
}

func New(devPath string, i2cAddress int) (*Bus, error) {
	d, err := i2c.Open(&i2c.Devfs{Dev: devPath}, i2cAddress)
	if err != nil {
		return nil, fmt.Errorf("opening i2c device: %w", err)
	}

	b := bme280.New(d)
	if err := b.Init(); err != nil {
		return nil, fmt.Errorf("initialization of bme280: %w", err)
	}

	return &Bus{
		device: d,
		bme280: b,
	}, nil
}

func (b *Bus) Read() (types.WeatherData, error) {
	t, p, h, err := b.bme280.EnvData()
	if err != nil {
		return types.WeatherData{}, fmt.Errorf("reading sensor data: %w", err)
	}
	return types.WeatherData{
		Temperature: t,
		Pressure:    p,
		Humidity:    h,
	}, nil

}

func (b *Bus) Close() error {
	return b.device.Close()
}
