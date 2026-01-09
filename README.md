# RaspiWeather

A lightweight weather monitoring service for Raspberry Pi that reads temperature, humidity, and pressure data from a BME280 sensor and exposes it via simple HTTP API and Prometheus metrics.

## Features

- Reads temperature, humidity, and pressure from BME280 sensor via I2C
- Exposes Prometheus metrics at `/metrics`
- Provides JSON endpoint at `/sensor-data`
- Configurable data collection interval
- Cross-platform builds for multiple architectures
- LCD1602 support

## Requirements

- Go 1.25.5 or later
- BME280 sensor connected via I2C
- Raspberry Pi (tested on old Model 1B)

## Building

```bash
# Build for current platform
make build

# Build for specific architecture
make build-linux-arm5
make build-linux-arm6
make build-linux-arm7
make build-linux-arm64
make build-linux-amd64

# Build for all architectures
make build-all

# Clean build artifacts
make clean
```

## Installation

1. Copy the binary to your Raspberry Pi:
   ```bash
   scp raspiweather-linux-arm7 pi@raspberrypi:/usr/bin/raspiweather
   ```

2. Make it executable (just in case):
   ```bash
   chmod +x /usr/bin/raspiweather
   ```

3. Enable I2C on your Raspberry Pi (if not already enabled):
   ```bash
   sudo raspi-config
   # Navigate to: Interfacing Options -> I2C -> Enable
   ```

## Usage

```bash
# Run with default settings (reads from /dev/i2c-1 every minute)
./raspiweather
```

### Command Line Options

```
$ raspiweather -h
Usage of raspiweather:
  -address string
        Address for HTTP Server (default ":9111")
  -bme280Addr int
        Address of bme280 (default 119)
  -bmeDevPath string
        Path to i2c bme device (default "/dev/i2c-1")
  -interval duration
        Interval of collecting sensors data (default 1m0s)
  -lcd
        Enable LCD1602
  -lcdAddr int
        Address of lcd1602 (default 39)
  -lcdBacklight
        Turn on LCD backlight
  -lcdCols int
        Number of LCD columns (default 16)
  -lcdDevPath string
        Path to i2c lcd device (default "/dev/i2c-1")
  -lcdRows int
        Number of LCD rows (default 2)
  -v    Show version and exit
```

## API Endpoints

### GET /sensor-data

Returns current sensor readings in JSON format:

```json
{
  "temperature": 25.0,
  "humidity": 44.0,
  "pressure": 1000.0,
}
```

### GET /metrics

Prometheus metrics endpoint with the following metrics:

- `sensor_temperature` - Current temperature in Celsius
- `sensor_humidity` - Current humidity in percent
- `sensor_pressure` - Current pressure in hPa

## Examples

The [examples](examples) directory contains configuration files and templates to help you get started:

- [examples/systemd/raspiweather.service](examples/systemd/raspiweather.service) - Systemd service file for running raspiweather as a system service. 

- [examples/grafana/dashboard.json](examples/grafana/dashboard.json) - Grafana dashboard configuration for visualizing temperature, humidity, and pressure metrics from Prometheus.

## Misc

There is a simple telegram bot for this project: https://github.com/RealFatCat/raspiweatherbot
